package model

import (
	"time"
)

type ModelConfig struct {
	ModelName         string
	PageTitle         string                            `json:"pageTitle"`
	DbTable           string                            `json:"dbTable"`
	SqlWhere          string                            `json:"sqlWhere"`
	Fields            []string                          `json:"fields"`
	OrderBy           string                            `json:"orderBy"`
	Headers           map[string]string                 `json:"headers"`
	Titles            map[string]string                 `json:"titles"`
	Classes           map[string]string                 `json:"classes"`
	DateTimeFields    map[string]string                 `json:"dateTimeFields"`
	RelatedData       map[string]string                 `json:"relatedData"`
	AddableFields     []string                          `json:"addableFields"`
	RequiredFields    []string                          `json:"requiredFields"`
	EditableFields    []string                          `json:"editableFields"`
	FieldEditor       map[string]string                 `json:"fieldsEditor"`
	NoZeroValueFields []string                          `json:"noZeroValueFields"`
	CountRelatedData  map[string]CountRelatedDataConfig `json:"countRelatedData"`
	Links             map[string]LinkConfig             `json:"links"`
	Parent            map[string]string                 `json:"parent"`
	DbTablePrimaryKey string
	ParentConfig      *ModelConfig
	FieldConfig       map[string]FieldParams
}

type CachedModelConfig struct {
	Config  *ModelConfig
	ModTime time.Time
}

type CountRelatedDataConfig struct {
	LocalFieldID  string `json:"localFieldID"`
	Table         string `json:"table"`
	TargetFieldID string `json:"targetFieldID"`
}

type LinkConfig struct {
	Template string `json:"template"`
}

type FieldParams struct {
	IsAddable   bool
	IsEditable  bool
	IsRequired  bool
	FieldEditor string
	Classes     string
}

type RenderTableCache struct {
	RelatedData map[string]string
}
