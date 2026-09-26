package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"youlai-gin/internal/codegen/model"
	"youlai-gin/pkg/errs"
	commonModel "youlai-gin/pkg/model"
)

// GenConfigRow gen_table 配置行（保存/回填代码生成配置）
type GenConfigRow struct {
	ID                int64  `gorm:"column:id"`
	TableName         string `gorm:"column:table_name"`
	ModuleName        string `gorm:"column:module_name"`
	PackageName       string `gorm:"column:package_name"`
	BusinessName      string `gorm:"column:business_name"`
	EntityName        string `gorm:"column:entity_name"`
	Author            string `gorm:"column:author"`
	ParentMenuID      *int64 `gorm:"column:parent_menu_id"`
	RemoveTablePrefix string `gorm:"column:remove_table_prefix"`
	PageType          string `gorm:"column:page_type"`
	IsDeleted         int    `gorm:"column:is_deleted"`
}

// GenColumn gen_table_column 字段配置行（读与覆盖写复用）
type GenColumn struct {
	ID            int64  `gorm:"column:id"`
	TableID       int64  `gorm:"column:table_id"`
	ColumnName    string `gorm:"column:column_name"`
	ColumnType    string `gorm:"column:column_type"`
	FieldName     string `gorm:"column:field_name"`
	FieldType     string `gorm:"column:field_type"`
	FieldSort     *int   `gorm:"column:field_sort"`
	FieldComment  string `gorm:"column:field_comment"`
	MaxLength     *int   `gorm:"column:max_length"`
	IsRequired    int    `gorm:"column:is_required"`
	IsShowInList  int    `gorm:"column:is_show_in_list"`
	IsShowInForm  int    `gorm:"column:is_show_in_form"`
	IsShowInQuery int    `gorm:"column:is_show_in_query"`
	QueryType     int    `gorm:"column:query_type"`
	FormType      int    `gorm:"column:form_type"`
	DictType      string `gorm:"column:dict_type"`
}

// TableColumn information_schema.COLUMNS 行（未配置时代码生成默认字段用）
type TableColumn struct {
	ColumnName      string `gorm:"column:columnName"`
	ColumnType      string `gorm:"column:columnType"`
	ColumnComment   string `gorm:"column:columnComment"`
	IsNullable      string `gorm:"column:isNullable"`
	MaxLength       *int   `gorm:"column:maxLength"`
	OrdinalPosition int    `gorm:"column:ordinalPosition"`
}

// Repository 代码生成数据访问层（gen_table / gen_table_column / information_schema 元数据）
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// TableListPage 分页查询可代码生成的数据表（含是否已配置标识，来自 information_schema）
func (r *Repository) TableListPage(ctx context.Context, keywords string, offset, limit int) (*commonModel.PagedData, error) {
	params := make([]interface{}, 0)
	where := "t.TABLE_SCHEMA = DATABASE() AND t.TABLE_NAME NOT IN ('gen_table','gen_table_column')"
	if keywords != "" {
		where += " AND t.TABLE_NAME LIKE ?"
		params = append(params, "%"+keywords+"%")
	}

	listSQL := fmt.Sprintf(`
SELECT
  t.TABLE_NAME AS tableName,
  t.TABLE_COMMENT AS tableComment,
  t.TABLE_COLLATION AS tableCollation,
  t.ENGINE AS engine,
  DATE_FORMAT(t.CREATE_TIME, '%%Y-%%m-%%d %%H:%%i:%%s') AS createTime,
  IF(c.id IS NULL, 0, 1) AS isConfigured
FROM information_schema.TABLES t
LEFT JOIN gen_table c
  ON c.table_name = t.TABLE_NAME AND c.is_deleted = 0
WHERE %s
ORDER BY t.CREATE_TIME DESC
LIMIT ? OFFSET ?`, where)

	var list []model.TableInfoVO
	if err := r.db.WithContext(ctx).Raw(listSQL, append(params, limit, offset)...).Scan(&list).Error; err != nil {
		return nil, errs.SystemError("查询数据表失败")
	}

	totalSQL := fmt.Sprintf(`SELECT COUNT(1) AS total FROM information_schema.TABLES t WHERE %s`, where)
	var total int64
	if err := r.db.WithContext(ctx).Raw(totalSQL, params...).Scan(&total).Error; err != nil {
		return nil, errs.SystemError("查询数据表失败")
	}

	return &commonModel.PagedData{List: list, Total: total}, nil
}

// GetGenTable 查询 gen_table 配置（未配置返回 found=false 且无错误）
func (r *Repository) GetGenTable(ctx context.Context, tableName string) (*GenConfigRow, bool, error) {
	var cfg GenConfigRow
	tx := r.db.WithContext(ctx).Table("gen_table").
		Where("table_name = ? AND is_deleted = 0", tableName).
		Limit(1).Find(&cfg)
	if tx.Error != nil {
		return nil, false, errs.SystemError("查询生成配置失败")
	}
	return &cfg, tx.RowsAffected > 0, nil
}

// GetGenColumns 查询指定配置的字段（按字段排序 asc）
func (r *Repository) GetGenColumns(ctx context.Context, tableID int64) ([]GenColumn, error) {
	var fields []GenColumn
	if err := r.db.WithContext(ctx).Table("gen_table_column").
		Where("table_id = ?", tableID).
		Order("field_sort ASC").
		Find(&fields).Error; err != nil {
		return nil, errs.SystemError("查询字段配置失败")
	}
	return fields, nil
}

// TableComment 查询数据表注释（information_schema，未配置时回填表名/业务名）
func (r *Repository) TableComment(ctx context.Context, tableName string) (string, error) {
	var comment string
	err := r.db.WithContext(ctx).Raw(`
SELECT TABLE_COMMENT AS tableComment FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? LIMIT 1`, tableName).Scan(&comment).Error
	return comment, err
}

// TableColumns 查询数据表字段元数据（information_schema，未配置时生成默认配置）
func (r *Repository) TableColumns(ctx context.Context, tableName string) ([]TableColumn, error) {
	var cols []TableColumn
	if err := r.db.WithContext(ctx).Raw(`
SELECT
  COLUMN_NAME AS columnName,
  DATA_TYPE AS columnType,
  COLUMN_COMMENT AS columnComment,
  IS_NULLABLE AS isNullable,
  CHARACTER_MAXIMUM_LENGTH AS maxLength,
  ORDINAL_POSITION AS ordinalPosition
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
ORDER BY ORDINAL_POSITION ASC`, tableName).Scan(&cols).Error; err != nil {
		return nil, errs.SystemError("查询表字段失败")
	}
	return cols, nil
}

// CreateGenTable 新增配置并返回自增ID
func (r *Repository) CreateGenTable(ctx context.Context, data map[string]interface{}) (int64, error) {
	if err := r.db.WithContext(ctx).Table("gen_table").Create(data).Error; err != nil {
		return 0, errs.SystemError("保存配置失败")
	}
	var row GenConfigRow
	if err := r.db.WithContext(ctx).Table("gen_table").
		Where("table_name = ?", data["table_name"]).First(&row).Error; err != nil {
		return 0, errs.SystemError("保存配置失败")
	}
	return row.ID, nil
}

// UpdateGenTable 更新配置
func (r *Repository) UpdateGenTable(ctx context.Context, id int64, updates map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Table("gen_table").
		Where("id = ?", id).Updates(updates).Error; err != nil {
		return errs.SystemError("保存配置失败")
	}
	return nil
}

// ReplaceGenColumns 覆盖写字段配置（先删后插）
func (r *Repository) ReplaceGenColumns(ctx context.Context, tableID int64, cols []GenColumn) error {
	if err := r.db.WithContext(ctx).Table("gen_table_column").
		Where("table_id = ?", tableID).Delete(&GenColumn{}).Error; err != nil {
		return errs.SystemError("保存字段配置失败")
	}
	for i := range cols {
		if err := r.db.WithContext(ctx).Table("gen_table_column").Create(&cols[i]).Error; err != nil {
			return errs.SystemError("保存字段配置失败")
		}
	}
	return nil
}

// DeleteGenConfig 逻辑删除配置并物理删除其字段（未配置视为已删除）
func (r *Repository) DeleteGenConfig(ctx context.Context, tableName string) error {
	var cfg GenConfigRow
	if err := r.db.WithContext(ctx).Table("gen_table").
		Where("table_name = ? AND is_deleted = 0", tableName).First(&cfg).Error; err != nil {
		return nil
	}

	if err := r.db.WithContext(ctx).Table("gen_table_column").
		Where("table_id = ?", cfg.ID).Delete(&GenColumn{}).Error; err != nil {
		return errs.SystemError("删除失败")
	}
	if err := r.db.WithContext(ctx).Table("gen_table").
		Where("id = ?", cfg.ID).
		Updates(map[string]interface{}{"is_deleted": 1, "update_time": time.Now()}).Error; err != nil {
		return errs.SystemError("删除失败")
	}
	return nil
}
