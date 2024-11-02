package wedyta

import "github.com/gin-gonic/gin"

type RenderDbTableConfig struct {
	AccessCheckFunc func(context *gin.Context, modelName, action, fieldName string) bool
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
	Table      string `json:"table"`
	ForeignKey string `json:"foreignKey"`
}

type LinkConfig struct {
	Template string `json:"template"`
}

type DbChanges struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"internal_id"`
	WebUserID      int    `gorm:"not null;default:0" json:"web_user_id"`
	ModelName      string `gorm:"not null" json:"model_name"`
	IdOfRecord     int    `gorm:"not null;default:0" json:"id_of_record"`
	DataFrom       string `gorm:"not null" json:"data_from"`
	DataTo         string `gorm:"not null" json:"data_to"`
	AddedTimestamp int64  `gorm:"autoCreateTime" json:"added_timestamp"`
}
