package common

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hsyan2008/hfw/common"
	"xorm.io/core"
)

type GenTool struct {
	targetDir   string
	packageName string
	tables      []*core.Table
	models      map[string]model
}

func NewGenTool() *GenTool {
	dir := Configs().TargetDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(common.GetAppPath(), dir)
	}
	if !common.IsExist(dir) {
		os.MkdirAll(dir, 0755)
	}
	return &GenTool{
		targetDir:   Configs().TargetDir,
		packageName: filepath.Base(dir),
		models:      make(map[string]model),
	}
}

func (genTool *GenTool) getDBMetas() (err error) {
	genTool.tables, err = DBMetas(Configs().Tables, Configs().ExcludeTables, Configs().TryComplete)
	if err != nil {
		return
	}

	return nil
}

func (genTool *GenTool) genModels() {
	for _, table := range genTool.tables {
		model := NewModel(table)
		genTool.models[model.TableName] = model
	}

	return
}

func (genTool *GenTool) genFile() (err error) {
	for _, model := range genTool.models {
		log.Println("start gen table:", model.TableName)

		//package
		str := fmt.Sprintln("package", genTool.packageName)

		//import
		if len(model.Imports) > 0 {
			str += fmt.Sprintln("import (")
			for _, i := range model.Imports {
				str += fmt.Sprintf(`"%s"%s`, i, "\n")
			}
			str += fmt.Sprintln(")")
		}

		//struct
		str += fmt.Sprintln("type", model.StructName, "struct {")
		for _, v := range model.Fields {
			str += fmt.Sprintln(v.FieldName, v.Type, v.Tag, v.Comment)
		}
		str += fmt.Sprintln("}")

		//func
		str += fmt.Sprintln("func (", model.StructName, ") TableName() string {")
		str += fmt.Sprintln(fmt.Sprintf("return `%s`", model.TableName))
		str += fmt.Sprintln("}")

		//format
		b, err := format.Source([]byte(str))
		if err != nil {
			return err
		}
		file := filepath.Join(genTool.targetDir, fmt.Sprintf("%s.go", model.TableName))
		err = ioutil.WriteFile(file, b, 0644)
		if err != nil {
			return err
		}
		log.Println("gen into file:", file)
	}

	return
}

func (genTool *GenTool) Gen() (err error) {
	if err = InitDb(); err != nil {
		return
	}
	log.Println("init db connect success!")

	if err = genTool.getDBMetas(); err != nil {
		return
	}
	log.Println("get tables info success!")

	genTool.genModels()
	log.Println("format tables info success!")

	log.Println("start generate model files...")
	if err = genTool.genFile(); err != nil {
		return
	}
	log.Println("generate complete!")

	return nil
}
