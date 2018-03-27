// +build ignore

package modelsvc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/NatLibFi/qvain-api/models"
)

type ModelSvc map[int]*ModelView

func NewFromFile(fn string) (*ModelSvc, error) {
	svc := make(ModelSvc)

	fh := os.Open(fn)
	defer fh.Close()
	if err != nil {
		return nil, err
	}
	parser := json.NewDecoder(fh)
	err := parser.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}
