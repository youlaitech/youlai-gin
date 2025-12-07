package excel

import (
	"fmt"
	"io"
	"reflect"
	"time"
	
	"github.com/xuri/excelize/v2"
)

// ExcelExporter Excel 导出器
type ExcelExporter struct {
	file      *excelize.File
	sheetName string
	row       int
}

// NewExcelExporter 创建 Excel 导出器
func NewExcelExporter(sheetName string) *ExcelExporter {
	file := excelize.NewFile()
	file.NewSheet(sheetName)
	file.DeleteSheet("Sheet1")
	return &ExcelExporter{
		file:      file,
		sheetName: sheetName,
		row:       1,
	}
}

// SetHeaders 设置表头
func (e *ExcelExporter) SetHeaders(headers []string) error {
	for i, header := range headers {
		cell := fmt.Sprintf("%s%d", numToCol(i), e.row)
		if err := e.file.SetCellValue(e.sheetName, cell, header); err != nil {
			return err
		}
	}
	
	// 设置表头样式
	headerStyle, _ := e.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	
	endCol := numToCol(len(headers) - 1)
	e.file.SetCellStyle(e.sheetName, fmt.Sprintf("A%d", e.row), fmt.Sprintf("%s%d", endCol, e.row), headerStyle)
	e.row++
	return nil
}

// AddRow 添加数据行
func (e *ExcelExporter) AddRow(values []interface{}) error {
	for i, value := range values {
		cell := fmt.Sprintf("%s%d", numToCol(i), e.row)
		if err := e.file.SetCellValue(e.sheetName, cell, value); err != nil {
			return err
		}
	}
	e.row++
	return nil
}

// Write 写入到 Writer
func (e *ExcelExporter) Write(w io.Writer) error {
	return e.file.Write(w)
}

// SaveAs 保存为文件
func (e *ExcelExporter) SaveAs(filename string) error {
	return e.file.SaveAs(filename)
}

// Close 关闭文件
func (e *ExcelExporter) Close() error {
	return e.file.Close()
}

// ExcelImporter Excel 导入器
type ExcelImporter struct {
	file      *excelize.File
	sheetName string
}

// NewExcelImporter 创建 Excel 导入器
func NewExcelImporter(reader io.Reader) (*ExcelImporter, error) {
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	sheetName := file.GetSheetName(0)
	return &ExcelImporter{file: file, sheetName: sheetName}, nil
}

// GetRows 获取所有行数据
func (i *ExcelImporter) GetRows() ([][]string, error) {
	return i.file.GetRows(i.sheetName)
}

// Close 关闭文件
func (i *ExcelImporter) Close() error {
	return i.file.Close()
}

// numToCol 数字转列名（0->A, 1->B, 26->AA）
func numToCol(n int) string {
	result := ""
	for n >= 0 {
		result = string(rune('A'+n%26)) + result
		n = n/26 - 1
		if n < 0 {
			break
		}
	}
	return result
}

// ExportToExcel 快速导出（泛型辅助函数）
func ExportToExcel(sheetName string, headers []string, data [][]interface{}) (*ExcelExporter, error) {
	exporter := NewExcelExporter(sheetName)
	if err := exporter.SetHeaders(headers); err != nil {
		return nil, err
	}
	for _, row := range data {
		if err := exporter.AddRow(row); err != nil {
			return nil, err
		}
	}
	return exporter, nil
}
