package wedyta

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Config struct {
	// Path to the folder where json configurations are located.
	// Default: config/wedyta
	ConfigDir string

	// The function must return true if the action on the specified table field is allowed.
	// It should be noted that in some cases the field may be empty when the access check occurs in the context of the entire table, and not a specific field.
	// It is recommended to place the function in such a way that it has access to the existing functions for checking authorization by the cookie of the main application, and this is the reason why the context is also passed to it.
	AccessCheckFunc func(context *gin.Context, modelName, fieldName, action string) bool

	// Gin template in which the content generated by the wedyta module will be placed
	Template string

	// A function that will add a list of variables and their values ​​that should be additionally filled in the template, for example, the username and the like.
	PrepareTemplateVariables func(context *gin.Context, modelName string, h gin.H)

	// HeadersTag default 'h2'
	HeadersTag string

	// BreadcrumbsRootName default 'Home'
	BreadcrumbsRootName string

	// BreadcrumbsRootUrl default '/'
	BreadcrumbsRootUrl string

	BeforeCreate func(context *gin.Context, db *gorm.DB, table string, id int64)
	BeforeUpdate func(context *gin.Context, db *gorm.DB, table string, id int64, field string)
	BeforeDelete func(context *gin.Context, db *gorm.DB, table string, id int64)
	AfterCreate  func(context *gin.Context, db *gorm.DB, table string, id int64)
	AfterUpdate  func(context *gin.Context, db *gorm.DB, table string, id int64, field string, valueBeforeUpdate string, valueAfterUpdate string)
	AfterDelete  func(context *gin.Context, db *gorm.DB, table string, id int64)
}

type modelConfig struct {
	PageTitle         string                            `json:"pageTitle"`
	DbTable           string                            `json:"dbTable"`
	SqlWhere          string                            `json:"sqlWhere"`
	Fields            []string                          `json:"fields"`
	OrderBy           string                            `json:"orderBy"`
	Headers           map[string]string                 `json:"headers"`
	Titles            map[string]string                 `json:"titles"`
	Classes           map[string]string                 `json:"classes"`
	RelatedData       map[string]string                 `json:"relatedData"`
	AddableFields     []string                          `json:"addableFields"`
	RequiredFields    []string                          `json:"requiredFields"`
	EditableFields    map[string]string                 `json:"editableFields"`
	NoZeroValueFields []string                          `json:"noZeroValueFields"`
	CountRelatedData  map[string]CountRelatedDataConfig `json:"countRelatedData"`
	Links             map[string]LinkConfig             `json:"links"`
	Parent            map[string]string                 `json:"parent"`
	ParentConfig      *modelConfig
}

type CountRelatedDataConfig struct {
	LocalFieldID  string `json:"localFieldID"`
	Table         string `json:"table"`
	TargetFieldID string `json:"targetFieldID"`
}

type LinkConfig struct {
	Template string `json:"template"`
}
