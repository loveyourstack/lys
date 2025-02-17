package coresetfunc

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Core setfunc"
	schemaName     string = "core"
	setFuncName    string = "setfunc"
	defaultOrderBy string = "text_val, int_val"
)

var setFuncUrlParamNames = []string{"p_text", "p_int", "p_inta"}

type Model struct {
	IntVal  int    `db:"int_val" json:"int_val"`
	TextVal string `db:"text_val" json:"text_val"`
}

var (
	meta lysmeta.Result
)

func init() {
	var err error
	meta, err = lysmeta.AnalyzeStructs(reflect.ValueOf(&Model{}).Elem())
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeStructs failed for %s.%s: %s", schemaName, setFuncName, err.Error())
	}
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) GetMeta() lysmeta.Result {
	return meta
}
func (s Store) GetName() string {
	return name
}
func (s Store) GetSetFuncUrlParamNames() []string {
	return setFuncUrlParamNames
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {

	// if it was sent, change the 3rd setFunc param from string to int array
	if len(params.SetFuncParamValues) == 3 {
		pIntA := []int{}
		pIntAStr := fmt.Sprintf("%v", params.SetFuncParamValues[2])
		pIntAStrA := strings.Split(pIntAStr, ",")
		for _, str := range pIntAStrA {
			pInt, err := strconv.Atoi(str)
			if err != nil {
				return nil, lyspg.TotalCount{}, fmt.Errorf("strconv.Atoi failed for value '%s': %w", str, err)
			}
			pIntA = append(pIntA, pInt)
		}
		params.SetFuncParamValues[2] = pIntA
	}

	return lyspg.Select[Model](ctx, s.Db, schemaName, "", setFuncName, defaultOrderBy, meta.DbTags, params)
}
