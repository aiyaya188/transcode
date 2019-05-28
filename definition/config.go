package definition
/*
import (
	"github.com/jinzhu/gorm"
)
*/

type Configs struct {
	ID     int    `gorm:"AUTO_INCREMENT"`
	Domain string `gorm:"size:255"`
}

type Cfgs struct {
	//gorm.Model	
	ID     int    `gorm:"AUTO_INCREMENT"`
	DomainNameSetting  string  `gorm:"size:255"`
	SubtitleSettings  string  `gorm:"size:255"`
	Colour  string  `gorm:"size:255"`
	SubtitleStatus    string   `gorm:"size:255"`
	FontSize		string     `gorm:"size:255"`
    StartingPosition string     `gorm:"size:255"`
}

type TranscodingConfgs struct {
	//gorm.Model	
	ID				int   `gorm:"AUTO_INCREMENT"`
	NumberOfTasks  string  `gorm:"size:255"`
	TranscodingFormat string  `gorm:"size:255"`
	LeftUpperWatermark  string  `gorm:"size:255"`
	RightUpperWatermark    string   `gorm:"size:255"`
	LeftLowerWatermark		string     `gorm:"size:255"`
	RightLowerWatermark		string     `gorm:"size:255"`
	WatermarkOrNot string     `gorm:"size:255"`
	FragmentationDuration    string     `gorm:"size:255"`
}


