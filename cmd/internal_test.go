package cmd

import (
	"errors"
	"testing"

	"github.com/bruin-data/bruin/pkg/pipeline"
)

func createEmployeeColumns() []pipeline.Column {
	return []pipeline.Column{
		{Name: "id", Type: "str", PrimaryKey: true},
		{Name: "name", Type: "str"},
		{Name: "age", Type: "int64"},
	}
}

func createJoinColumns() []pipeline.Column {
	return []pipeline.Column{
		{Name: "a", Type: "str"},
		{Name: "b", Type: "int64"},
		{Name: "c", Type: "str"},
		{Name: "b2", Type: "int64"},
		{Name: "c2", Type: "str"},
	}
}

const complexJoinQuery = `
	with t1 as (
		select *
		from table1
		join table2
			using(a)
	),
	t2 as (
		select *
		from table2
		left join table1
			using(a)
	)
	select t1.*, t2.b as b2, t2.c as c2
	from t1
	join t2
		using(a)
`

func createComplexJoinPipeline() *pipeline.Pipeline {
	return &pipeline.Pipeline{
		Assets: []*pipeline.Asset{
			{
				Name: "table1",
				Columns: []pipeline.Column{
					{Name: "a", Type: "str"},
					{Name: "b", Type: "int64"},
				},
			},
			{
				Name: "table2",
				Columns: []pipeline.Column{
					{Name: "a", Type: "str"},
					{Name: "c", Type: "str"},
				},
			},
		},
	}
}

func TestInternalParse_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		pipeline     *pipeline.Pipeline
		beforeAssets *pipeline.Asset
		afterAssets  *pipeline.Asset
		wantCount    int
		wantColumns  []pipeline.Column
		want         error
	}{
		{
			name: "simple select all query",
			pipeline: &pipeline.Pipeline{
				Assets: []*pipeline.Asset{
					{
						Name:    "employees",
						Columns: createEmployeeColumns(),
					},
				},
			},
			beforeAssets: &pipeline.Asset{
				Name: "example",
				ExecutableFile: pipeline.ExecutableFile{
					Content: "select * from employees",
				},
				Upstreams: []pipeline.Upstream{{Value: "employees"}},
			},
			afterAssets: &pipeline.Asset{
				Name: "example",
				ExecutableFile: pipeline.ExecutableFile{
					Content: "select * from employees",
				},
				Columns:   createEmployeeColumns(),
				Upstreams: []pipeline.Upstream{{Value: "employees"}},
			},
			wantCount:   3,
			wantColumns: createEmployeeColumns(),
			want:        nil,
		},
		{
			name: "simple select all query wihtout upstream",
			pipeline: &pipeline.Pipeline{
				Assets: []*pipeline.Asset{
					{
						Name:    "employees",
						Columns: createEmployeeColumns(),
					},
				},
			},
			beforeAssets: &pipeline.Asset{
				Name: "example",
				ExecutableFile: pipeline.ExecutableFile{
					Content: "select * from employees",
				},
				Upstreams: []pipeline.Upstream{},
			},
			afterAssets: &pipeline.Asset{
				Name: "example",
				ExecutableFile: pipeline.ExecutableFile{
					Content: "select * from employees",
				},
				Columns:   []pipeline.Column{},
				Upstreams: []pipeline.Upstream{},
			},
			wantCount:   0,
			wantColumns: []pipeline.Column{},
			want:        nil,
		},
		{
			name:     "complex join query",
			pipeline: createComplexJoinPipeline(),
			beforeAssets: &pipeline.Asset{
				Name: "example",
				ExecutableFile: pipeline.ExecutableFile{
					Content: complexJoinQuery,
				},
				Columns:   []pipeline.Column{},
				Upstreams: []pipeline.Upstream{{Value: "table1"}, {Value: "table2"}},
			},
			afterAssets: &pipeline.Asset{
				Name: "example",
				ExecutableFile: pipeline.ExecutableFile{
					Content: complexJoinQuery,
				},
				Columns:   createJoinColumns(),
				Upstreams: []pipeline.Upstream{{Value: "table1"}, {Value: "table2"}},
			},
			wantCount:   5,
			wantColumns: createJoinColumns(),
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ParseLineage(tt.pipeline, tt.beforeAssets)
			if !errors.Is(err, tt.want) {
				t.Errorf("ParseLineage() error = %v, want %v", err, tt.want)
			}

			if tt.afterAssets != nil {
				assertColumns(t, tt.afterAssets.Columns, tt.wantColumns, tt.wantCount)
			}
		})
	}
}

func TestParseLineageRecursively(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pipeline *pipeline.Pipeline
		after    *pipeline.Pipeline
		want     error
	}{
		{
			name: "successful recursive lineage parsing",
			pipeline: &pipeline.Pipeline{
				Assets: []*pipeline.Asset{
					{
						Name: "table1",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT * FROM table2",
						},
						Upstreams: []pipeline.Upstream{{Value: "table2"}},
					},
					{
						Name: "table2",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT * FROM table3",
						},
						Upstreams: []pipeline.Upstream{{Value: "table3"}},
					},
					{
						Name: "table3",
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "age", Type: "int64"},
						},
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT id,name,age FROM table3",
						},
					},
				},
			},
			after: &pipeline.Pipeline{
				Assets: []*pipeline.Asset{
					{
						Name: "table1",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT * FROM table2",
						},
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "age", Type: "int64"},
						},
						Upstreams: []pipeline.Upstream{{Value: "table2"}},
					},
					{
						Name: "table2",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT * FROM table3",
						},
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "age", Type: "int64"},
						},
						Upstreams: []pipeline.Upstream{{Value: "table3"}},
					},
					{
						Name: "table3",
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "age", Type: "int64"},
						},
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT id,name,age FROM table3",
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "successful recursive lineage parsing",
			pipeline: &pipeline.Pipeline{
				Assets: []*pipeline.Asset{
					{
						Name: "table1",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT name FROM table2",
						},
						Upstreams: []pipeline.Upstream{{Value: "table2"}},
					},
					{
						Name: "table2",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT id, age FROM table3",
						},
						Upstreams: []pipeline.Upstream{{Value: "table3"}},
					},
					{
						Name: "table3",
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "age", Type: "int64"},
						},
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT id,name,age FROM table3",
						},
					},
				},
			},
			after: &pipeline.Pipeline{
				Assets: []*pipeline.Asset{
					{
						Name: "table1",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT name FROM table2",
						},
						Columns: []pipeline.Column{
							{Name: "name", Type: "str"},
						},
						Upstreams: []pipeline.Upstream{{Value: "table2"}},
					},
					{
						Name: "table2",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT id, age FROM table3",
						},
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "age", Type: "int64"},
						},
						Upstreams: []pipeline.Upstream{{Value: "table3"}},
					},
					{
						Name: "table3",
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "age", Type: "int64"},
						},
						ExecutableFile: pipeline.ExecutableFile{
							Content: "SELECT id,name,age FROM table3",
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "successful recursive lineage parsing with joins",
			pipeline: &pipeline.Pipeline{
				Assets: []*pipeline.Asset{
					{
						Name: "analytics",
						ExecutableFile: pipeline.ExecutableFile{
							Content: `
select 
    a.name, 
    b.country 
from people a 
join country b on a.id = b.id;`,
						},
						Upstreams: []pipeline.Upstream{{Value: "country"}, {Value: "people"}},
					},
					{
						Name: "country",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "select id, country from users;",
						},
						Upstreams: []pipeline.Upstream{{Value: "users"}},
					},
					{
						Name: "people",
						ExecutableFile: pipeline.ExecutableFile{
							Content: `
select 
    id, 
    name, 
    last_name,
    now() as current_timestamp 
from users;`,
						},
						Upstreams: []pipeline.Upstream{{Value: "users"}},
					},
					{
						Name: "users",
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "last_name", Type: "str"},
							{Name: "country", Type: "str"},
							{Name: "created_at", Type: "timestamp"},
						},
						ExecutableFile: pipeline.ExecutableFile{
							Content: "select *  from user_data;",
						},
					},
				},
			},
			after: &pipeline.Pipeline{
				Assets: []*pipeline.Asset{
					{
						Name: "analytics",
						Columns: []pipeline.Column{
							{Name: "name", Type: "str"},
							{Name: "country", Type: "str"},
						},
						ExecutableFile: pipeline.ExecutableFile{
							Content: `
select 
    a.name, 
    b.country 
from people a 
join country b on a.id = b.id;`,
						},
						Upstreams: []pipeline.Upstream{{Value: "country"}, {Value: "people"}},
					},
					{
						Name: "country",
						ExecutableFile: pipeline.ExecutableFile{
							Content: "select id, country from users;",
						},
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "country", Type: "str"},
						},
						Upstreams: []pipeline.Upstream{{Value: "users"}},
					},
					{
						Name: "people",
						ExecutableFile: pipeline.ExecutableFile{
							Content: `
select 
    id, 
    name, 
    last_name,
    now() as current_timestamp 
from users;`,
						},
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "last_name", Type: "str"},
						},
						Upstreams: []pipeline.Upstream{{Value: "users"}},
					},
					{
						Name: "users",
						Columns: []pipeline.Column{
							{Name: "id", Type: "int64"},
							{Name: "name", Type: "str"},
							{Name: "last_name", Type: "str"},
							{Name: "country", Type: "str"},
							{Name: "created_at", Type: "timestamp"},
						},
						ExecutableFile: pipeline.ExecutableFile{
							Content: "select *  from user_data;",
						},
					},
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, asset := range tt.pipeline.Assets {
				err := parseLineageRecursive(tt.pipeline, asset)
				if tt.want == nil {
					if err != nil {
						t.Errorf("parseLineageRecursive() error = %v, want nil", err)
					}
				} else {
					if err == nil || err.Error() != tt.want.Error() {
						t.Errorf("parseLineageRecursive() error = %v, want %v", err, tt.want)
					}
				}

				for _, asset := range tt.after.Assets {
					assetFound := tt.pipeline.GetAssetByName(asset.Name)
					if assetFound == nil {
						t.Errorf("Asset %s not found in pipeline", asset.Name)
					}
				}
			}
		})
	}
}

func assertColumns(t *testing.T, got, want []pipeline.Column, wantCount int) {
	t.Helper()

	if len(got) != wantCount {
		t.Errorf("Column count mismatch: got %d, want %d", len(got), wantCount)
	}

	columnMap := make(map[string]pipeline.Column)
	for _, col := range want {
		columnMap[col.Name] = col
	}

	for _, col := range got {
		wantCol, exists := columnMap[col.Name]
		if !exists {
			t.Errorf("Unexpected column %s found", col.Name)
			continue
		}

		if col.Type != wantCol.Type {
			t.Errorf("Column %s type mismatch: got %s, want %s", col.Name, col.Type, wantCol.Type)
		}
		if col.PrimaryKey != wantCol.PrimaryKey {
			t.Errorf("Column %s primary key mismatch: got %v, want %v", col.Name, col.PrimaryKey, wantCol.PrimaryKey)
		}
	}
}
