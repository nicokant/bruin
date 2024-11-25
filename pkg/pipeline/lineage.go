package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bruin-data/bruin/pkg/dialect"
	"github.com/bruin-data/bruin/pkg/sqlparser"
)

type LineageExtractor struct {
	Pipeline       *Pipeline
	columnMetadata sqlparser.Schema
}

// NewLineageExtractor creates a new LineageExtractor instance.
func NewLineageExtractor(pipeline *Pipeline) *LineageExtractor {
	return &LineageExtractor{
		Pipeline:       pipeline,
		columnMetadata: make(sqlparser.Schema),
	}
}

// TableSchema extracts the table schema from the assets and stores it in the columnMetadata map.
func (p *LineageExtractor) TableSchema() error {
	for _, foundAsset := range p.Pipeline.Assets {
		if len(foundAsset.Columns) > 0 {
			p.columnMetadata[foundAsset.Name] = makeColumnMap(foundAsset.Columns)
		}
	}
	return nil
}

// ParseLineageRecursive processes the lineage of an asset and its upstream dependencies recursively.
func (p *LineageExtractor) ColumnLineage(asset *Asset) error {
	if asset == nil {
		return nil
	}

	if len(asset.Columns) > 0 {
		return nil
	}

	for _, upstream := range asset.Upstreams {
		upstreamAsset := p.Pipeline.GetAssetByName(upstream.Value)
		if upstreamAsset == nil {
			continue
		}
		_ = p.ColumnLineage(upstreamAsset)
	}

	_ = p.parseLineage(asset)

	return nil
}

// ParseLineage analyzes the column lineage for a given asset within a
// It traces column relationships between the asset and its upstream dependencies.
func (p *LineageExtractor) parseLineage(asset *Asset) error {
	if asset == nil {
		return errors.New("invalid arguments: asset and pipeline cannot be nil")
	}

	dialect, err := dialect.GetDialectByAssetType(string(asset.Type))
	if err != nil {
		return err
	}

	parser, err := sqlparser.NewSQLParser()
	if err != nil {
		return fmt.Errorf("failed to create SQL parser: %w", err)
	}

	if err := parser.Start(); err != nil {
		return fmt.Errorf("failed to start SQL parser: %w", err)
	}

	lineage, err := parser.ColumnLineage(asset.ExecutableFile.Content, dialect, p.columnMetadata)
	if err != nil {
		return fmt.Errorf("failed to parse column lineage: %w", err)
	}
	return p.processLineageColumns(asset, lineage)
}

func (p *LineageExtractor) processLineageColumns(asset *Asset, lineage *sqlparser.Lineage) error {
	if lineage == nil {
		return nil
	}

	if asset == nil {
		return errors.New("asset cannot be nil")
	}

	for _, lineageCol := range lineage.Columns {
		if lineageCol.Name == "*" {
			for _, upstream := range lineageCol.Upstream {
				upstreamAsset := p.Pipeline.GetAssetByName(upstream.Table)
				if upstreamAsset == nil {
					continue
				}

				// If upstream column is *, copy all columns from upstream asset
				if upstream.Column == "*" {
					for _, upstreamCol := range upstreamAsset.Columns {
						if err := p.addColumnToAsset(asset, upstreamCol.Name, upstreamAsset, &upstreamCol); err != nil {
							return err
						}
					}
					continue
				}
			}
			continue
		}

		for _, upstream := range lineageCol.Upstream {
			if upstream.Column == "*" {
				continue
			}
			if upstream.Table == asset.Name {
				continue
			}
			upstreamAsset := p.Pipeline.GetAssetByName(upstream.Table)
			if upstreamAsset == nil {
				if err := p.addColumnToAsset(asset, lineageCol.Name, nil, &Column{
					Name:   upstream.Column,
					Type:   strings.ToLower(upstream.Table),
					Checks: []ColumnCheck{},
					Upstreams: []*UpstreamColumn{
						{
							Asset:  strings.ToLower(upstream.Table),
							Column: upstream.Column,
							Table:  strings.ToLower(upstream.Table),
						},
					},
				}); err != nil {
					return err
				}
				continue
			}

			upstreamCol := upstreamAsset.GetColumnWithName(upstream.Column)
			if upstreamCol == nil {
				upstreamCol = &Column{
					Name:   upstream.Column,
					Type:   upstream.Table,
					Checks: []ColumnCheck{},
					Upstreams: []*UpstreamColumn{
						{
							Asset:  upstreamAsset.Name,
							Column: upstream.Table,
							Table:  upstreamAsset.Name,
						},
					},
				}
			}
			if err := p.addColumnToAsset(asset, lineageCol.Name, upstreamAsset, upstreamCol); err != nil {
				return err
			}
		}
	}
	return nil
}

// addColumnToAsset adds a new column to the asset based on upstream information.
func (p *LineageExtractor) addColumnToAsset(asset *Asset, colName string, upstreamAsset *Asset, upstreamCol *Column) error {
	if asset == nil || upstreamCol == nil || colName == "" {
		return errors.New("invalid arguments: all parameters must be non-nil and colName must not be empty")
	}

	if colName == "*" {
		return nil
	}

	newCol := Column{
		Name:       colName,
		PrimaryKey: false,
		Type:       upstreamCol.Type,
		Checks:     []ColumnCheck{},
		// Description:     upstreamCol.Description,
		EntityAttribute: upstreamCol.EntityAttribute,
		Upstreams:       []*UpstreamColumn{},
		UpdateOnMerge:   upstreamCol.UpdateOnMerge,
	}

	if upstreamAsset == nil {
		newCol = *upstreamCol
	}

	col := asset.GetColumnWithName(colName)

	if col != nil {
		if upstreamAsset != nil {
			newUpstream := &UpstreamColumn{
				Asset:  upstreamAsset.Name,
				Column: upstreamCol.Name,
				Table:  upstreamAsset.Name,
			}
			for i, existing := range asset.Columns {
				if existing.Name == colName {
					asset.Columns[i].Upstreams = append(asset.Columns[i].Upstreams, newUpstream)
					return nil
				}
			}
		}
		return nil
	}

	if upstreamAsset != nil {
		newCol.Upstreams = append(newCol.Upstreams, &UpstreamColumn{
			Asset:  upstreamAsset.Name,
			Column: upstreamCol.Name,
			Table:  upstreamAsset.Name,
		})
	}
	asset.Columns = append(asset.Columns, newCol)
	return nil
}

// makeColumnMap creates a map of column names to their types from a slice of columns.
func makeColumnMap(columns []Column) map[string]string {
	if len(columns) == 0 {
		return make(map[string]string)
	}

	columnMap := make(map[string]string, len(columns))
	for _, col := range columns {
		if col.Name != "" {
			columnMap[col.Name] = col.Type
		}
	}
	return columnMap
}
