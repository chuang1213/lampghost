package difftable

import (
	"encoding/json"
	"io"
	"os"
)

// struct DiffTable represents a content of difficult table
type DiffTable struct {
	Artist   string
	Comment  string
	Level    string
	Lr2BmsId string `json:"lr2_bmdid"`
	Md5      string
	NameDiff string
	Title    string
	Url      string `json:"url"`
	UrlDiff  string `json:"url_diff"`

	// Actually doesn't exist, must be merged from elsewhere
	Sha256    string `json:"-"`
	Lamp      int32  `json:"-"`
	GhostLamp int32  `json:"-"`
}

// Read Difficult Table content from disk.
func ReadDiffTable(filePath string) ([]DiffTable, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var local []DiffTable
	err = json.Unmarshal(body, &local)
	if err != nil {
		return nil, err
	}
	return local, nil
}

// Read difficult table content, and return a group split by level
func ReadDiffTableLevelMap(filePath string) (map[string][]DiffTable, error) {
	rawArr, err := ReadDiffTable(filePath)
	if err != nil {
		return nil, err
	}

	ret := make(map[string][]DiffTable)
	for _, v := range rawArr {
		if _, ok := ret[v.Level]; !ok {
			ret[v.Level] = make([]DiffTable, 0)
		}
		ret[v.Level] = append(ret[v.Level], v)
	}
	return ret, nil
}
