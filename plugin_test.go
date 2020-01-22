package nested_test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kirinse/gorm-nested"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"os"
	"testing"
)

var dbName = fmt.Sprintf("test_%d.db", rand.Int())

type PluginTestSuite struct {
	suite.Suite
	db *gorm.DB
}

type Taxon struct {
	ID       uint `gorm:"primary_key"`
	Name     string
	ParentID uint
	Parent   *Taxon `gorm:"association_autoupdate:false"`
	Lft      int    `gorm-nested:"left"`
	Rgt      int    `gorm-nested:"right"`
	Level    int    `gorm-nested:"level"`
}

func (t Taxon) GetParentID() interface{} {
	return t.ParentID
}

func (t Taxon) GetParent() nested.Interface {
	return t.Parent
}

func (suite *PluginTestSuite) SetupTest() {
	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		panic(fmt.Errorf("setup test: %s", err))
	}
	db = db.Debug()
	suite.db = db
	suite.db.AutoMigrate(&Taxon{})

	_, err = nested.Register(suite.db)
	if err != nil {
		panic(err)
	}
}

func (suite *PluginTestSuite) TearDownTest() {
	if err := suite.db.Close(); err != nil {
		panic(fmt.Errorf("tear down test: %s", err))
	}

	if err := os.Remove(dbName); err != nil {
		panic(fmt.Errorf("tear down test: %s", err))
	}
}

func (suite *PluginTestSuite) TestAddRoot() {
	root1 := Taxon{
		Name: "Root1",
	}
	suite.db.Create(&root1)

	root2 := Taxon{
		Name: "Root2",
	}
	suite.db.Create(&root2)

	root3 := Taxon{
		Name: "Root3",
	}
	suite.db.Create(&root3)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 3)
	assert.Equal(suite.T(), 1, taxons[0].Lft)
	assert.Equal(suite.T(), 2, taxons[0].Rgt)

	assert.Equal(suite.T(), 3, taxons[1].Lft)
	assert.Equal(suite.T(), 4, taxons[1].Rgt)

	assert.Equal(suite.T(), 5, taxons[2].Lft)
	assert.Equal(suite.T(), 6, taxons[2].Rgt)
}

func (suite *PluginTestSuite) TestInsertEntireTree() {
	node := Taxon{
		Name: "Tube",
		Parent: &Taxon{
			Name: "Television",
			Parent: &Taxon{
				Name: "Electronics",
			},
		},
	}
	suite.db.Create(&node)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 3)
	assert.Equal(suite.T(), 1, taxons[0].Lft)
	assert.Equal(suite.T(), 6, taxons[0].Rgt)

	assert.Equal(suite.T(), 2, taxons[1].Lft)
	assert.Equal(suite.T(), 5, taxons[1].Rgt)

	assert.Equal(suite.T(), 3, taxons[2].Lft)
	assert.Equal(suite.T(), 4, taxons[2].Rgt)
}

func (suite *PluginTestSuite) TestInsertNodeByNode() {
	television := Taxon{
		Name: "Television",
		Parent: &Taxon{
			Name: "Electronics",
		},
	}
	suite.db.Create(&television)

	tube := Taxon{
		Name:   "Tube",
		Parent: &television,
	}
	suite.db.Create(&tube)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 3)
	assert.Equal(suite.T(), 1, taxons[0].Lft)
	assert.Equal(suite.T(), 6, taxons[0].Rgt)
	assert.Equal(suite.T(), 0, taxons[0].Level)

	assert.Equal(suite.T(), 2, taxons[1].Lft)
	assert.Equal(suite.T(), 5, taxons[1].Rgt)
	assert.Equal(suite.T(), 1, taxons[1].Level)

	assert.Equal(suite.T(), 3, taxons[2].Lft)
	assert.Equal(suite.T(), 4, taxons[2].Rgt)
	assert.Equal(suite.T(), 2, taxons[2].Level)
}

func (suite *PluginTestSuite) TestDeleteNode() {
	suite.createTree()

	television := Taxon{}
	suite.db.First(&television, "name = 'Television'")
	suite.db.Delete(&television)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 7)
	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.Lft)
			assert.Equal(suite.T(), 14, taxon.Rgt)
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 2, taxon.Lft)
			assert.Equal(suite.T(), 3, taxon.Rgt)
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 4, taxon.Lft)
			assert.Equal(suite.T(), 13, taxon.Rgt)
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 5, taxon.Lft)
			assert.Equal(suite.T(), 8, taxon.Rgt)
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 6, taxon.Lft)
			assert.Equal(suite.T(), 7, taxon.Rgt)
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 9, taxon.Lft)
			assert.Equal(suite.T(), 10, taxon.Rgt)
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 11, taxon.Lft)
			assert.Equal(suite.T(), 12, taxon.Rgt)
			count++
			break
		}
	}

	var portableElectronics Taxon
	suite.db.First(&portableElectronics, "name = 'Portable Electronics'")
	suite.db.Delete(&portableElectronics)

	taxons = []Taxon{}
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 2)
	count = 0
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.Lft)
			assert.Equal(suite.T(), 4, taxon.Rgt)
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 2, taxon.Lft)
			assert.Equal(suite.T(), 3, taxon.Rgt)
			count++
			break
		}
	}

	var gameConsoles Taxon
	suite.db.First(&gameConsoles, "name = 'Game Consoles'")
	suite.db.Delete(&gameConsoles)
	taxons = []Taxon{}
	suite.db.Find(&taxons)

	assert.Equal(suite.T(), 1, taxons[0].Lft)
	assert.Equal(suite.T(), 2, taxons[0].Rgt)
}

func (suite *PluginTestSuite) TestMoveNodeToLeft() {
	suite.createTree()

	var portableElectronics Taxon
	var lcd Taxon

	assert.False(suite.T(), suite.db.First(&portableElectronics, "name = 'Portable Electronics'").RecordNotFound())
	assert.False(suite.T(), suite.db.First(&lcd, "name = 'LCD'").RecordNotFound())

	portableElectronics.Parent = &lcd
	portableElectronics.ParentID = lcd.ID

	suite.db.Save(&portableElectronics)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 11)

	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.Lft)
			assert.Equal(suite.T(), 22, taxon.Rgt)
			assert.Equal(suite.T(), 0, taxon.Level)
			count++
			break
		case "Television":
			assert.Equal(suite.T(), 2, taxon.Lft)
			assert.Equal(suite.T(), 19, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 20, taxon.Lft)
			assert.Equal(suite.T(), 21, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "Tube":
			assert.Equal(suite.T(), 3, taxon.Lft)
			assert.Equal(suite.T(), 4, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "LCD":
			assert.Equal(suite.T(), 5, taxon.Lft)
			assert.Equal(suite.T(), 16, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Plasma":
			assert.Equal(suite.T(), 17, taxon.Lft)
			assert.Equal(suite.T(), 18, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 6, taxon.Lft)
			assert.Equal(suite.T(), 15, taxon.Rgt)
			assert.Equal(suite.T(), 3, taxon.Level)
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 7, taxon.Lft)
			assert.Equal(suite.T(), 10, taxon.Rgt)
			assert.Equal(suite.T(), 4, taxon.Level)
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 8, taxon.Lft)
			assert.Equal(suite.T(), 9, taxon.Rgt)
			assert.Equal(suite.T(), 5, taxon.Level)
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 11, taxon.Lft)
			assert.Equal(suite.T(), 12, taxon.Rgt)
			assert.Equal(suite.T(), 4, taxon.Level)
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 13, taxon.Lft)
			assert.Equal(suite.T(), 14, taxon.Rgt)
			assert.Equal(suite.T(), 4, taxon.Level)
			count++
			break
		}
	}

	assert.Equal(suite.T(), len(taxons), count)
}

func (suite *PluginTestSuite) TestMoveNodeToRight() {
	suite.createTree()

	var mp3 Taxon
	var lcd Taxon

	assert.False(suite.T(), suite.db.First(&mp3, "name = 'MP3'").RecordNotFound())
	assert.False(suite.T(), suite.db.First(&lcd, "name = 'LCD'").RecordNotFound())

	lcd.Parent = &mp3
	lcd.ParentID = mp3.ID

	suite.db.Save(&lcd)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 11)

	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.Lft)
			assert.Equal(suite.T(), 22, taxon.Rgt)
			assert.Equal(suite.T(), 0, taxon.Level)
			count++
			break
		case "Television":
			assert.Equal(suite.T(), 2, taxon.Lft)
			assert.Equal(suite.T(), 7, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 8, taxon.Lft)
			assert.Equal(suite.T(), 9, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "Tube":
			assert.Equal(suite.T(), 3, taxon.Lft)
			assert.Equal(suite.T(), 4, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Plasma":
			assert.Equal(suite.T(), 5, taxon.Lft)
			assert.Equal(suite.T(), 6, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 10, taxon.Lft)
			assert.Equal(suite.T(), 21, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 11, taxon.Lft)
			assert.Equal(suite.T(), 16, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 12, taxon.Lft)
			assert.Equal(suite.T(), 13, taxon.Rgt)
			assert.Equal(suite.T(), 3, taxon.Level)
			count++
			break
		case "LCD":
			assert.Equal(suite.T(), 16, taxon.Lft)
			assert.Equal(suite.T(), 17, taxon.Rgt)
			assert.Equal(suite.T(), 3, taxon.Level)
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 17, taxon.Lft)
			assert.Equal(suite.T(), 18, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 19, taxon.Lft)
			assert.Equal(suite.T(), 20, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		}
	}

	assert.Equal(suite.T(), len(taxons), count)
}

func (suite *PluginTestSuite) TestChildNodeBecomesRoot() {
	suite.createTree()

	var mp3 Taxon

	assert.False(suite.T(), suite.db.First(&mp3, "name = 'MP3'").RecordNotFound())

	mp3.Parent = nil
	mp3.ParentID = 0

	suite.db.Save(&mp3)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 11)

	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.Lft)
			assert.Equal(suite.T(), 18, taxon.Rgt)
			assert.Equal(suite.T(), 0, taxon.Level)
			count++
			break
		case "Television":
			assert.Equal(suite.T(), 2, taxon.Lft)
			assert.Equal(suite.T(), 9, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 10, taxon.Lft)
			assert.Equal(suite.T(), 11, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "Tube":
			assert.Equal(suite.T(), 3, taxon.Lft)
			assert.Equal(suite.T(), 4, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "LCD":
			assert.Equal(suite.T(), 5, taxon.Lft)
			assert.Equal(suite.T(), 6, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Plasma":
			assert.Equal(suite.T(), 7, taxon.Lft)
			assert.Equal(suite.T(), 8, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 12, taxon.Lft)
			assert.Equal(suite.T(), 17, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 13, taxon.Lft)
			assert.Equal(suite.T(), 14, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 15, taxon.Lft)
			assert.Equal(suite.T(), 16, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 19, taxon.Lft)
			assert.Equal(suite.T(), 22, taxon.Rgt)
			assert.Equal(suite.T(), 0, taxon.Level)
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 20, taxon.Lft)
			assert.Equal(suite.T(), 21, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		}
	}

	assert.Equal(suite.T(), len(taxons), count)
}

func (suite *PluginTestSuite) TestAutoUpdateParentAssociation() {
	electronics := Taxon{
		Name: "Electronics",
	}

	television := Taxon{
		Name:   "Television",
		Parent: &electronics,
	}
	gameConsoles := Taxon{
		Name:   "Game Consoles",
		Parent: &electronics,
	}
	portableElectronics := Taxon{
		Name:   "Portable Electronics",
		Parent: &electronics,
	}

	tube := Taxon{
		Name:   "Tube",
		Parent: &television,
	}
	lcd := Taxon{
		Name:   "LCD",
		Parent: &television,
	}
	plasma := Taxon{
		Name:   "Plasma",
		Parent: &television,
	}

	mp3 := Taxon{
		Name:   "MP3",
		Parent: &portableElectronics,
	}

	cdPlayer := Taxon{
		Name:   "CD Player",
		Parent: &portableElectronics,
	}

	radio := Taxon{
		Name:   "Radio",
		Parent: &portableElectronics,
	}

	flash := Taxon{
		Name:   "Flash",
		Parent: &mp3,
	}

	suite.db.Save(&television)
	suite.db.Save(&gameConsoles)
	suite.db.Save(&tube)
	suite.db.Save(&lcd)
	suite.db.Save(&plasma)
	suite.db.Save(&flash)
	suite.db.Save(&cdPlayer)
	suite.db.Save(&radio)

	assert.Equal(suite.T(), 1, electronics.Lft)
	assert.Equal(suite.T(), 22, electronics.Rgt)
	assert.Equal(suite.T(), 0, electronics.Level)

	assert.Equal(suite.T(), 2, television.Lft)
	assert.Equal(suite.T(), 9, television.Rgt)
	assert.Equal(suite.T(), 1, television.Level)

	assert.Equal(suite.T(), 3, tube.Lft)
	assert.Equal(suite.T(), 4, tube.Rgt)
	assert.Equal(suite.T(), 2, tube.Level)

	assert.Equal(suite.T(), 5, lcd.Lft)
	assert.Equal(suite.T(), 6, lcd.Rgt)
	assert.Equal(suite.T(), 2, lcd.Level)

	assert.Equal(suite.T(), 7, plasma.Lft)
	assert.Equal(suite.T(), 8, plasma.Rgt)
	assert.Equal(suite.T(), 2, plasma.Level)

	//assert.Equal(suite.T(), 10, gameConsoles.Lft)
	//assert.Equal(suite.T(), 11, gameConsoles.Rgt)
	//assert.Equal(suite.T(), 1, gameConsoles.Level)

	assert.Equal(suite.T(), 12, portableElectronics.Lft)
	assert.Equal(suite.T(), 21, portableElectronics.Rgt)
	assert.Equal(suite.T(), 1, portableElectronics.Level)

	assert.Equal(suite.T(), 13, mp3.Lft)
	assert.Equal(suite.T(), 16, mp3.Rgt)
	assert.Equal(suite.T(), 2, mp3.Level)

	assert.Equal(suite.T(), 17, cdPlayer.Lft)
	assert.Equal(suite.T(), 18, cdPlayer.Rgt)
	assert.Equal(suite.T(), 2, cdPlayer.Level)

	assert.Equal(suite.T(), 19, radio.Lft)
	assert.Equal(suite.T(), 20, radio.Rgt)
	assert.Equal(suite.T(), 2, radio.Level)

	assert.Equal(suite.T(), 14, flash.Lft)
	assert.Equal(suite.T(), 15, flash.Rgt)
	assert.Equal(suite.T(), 3, flash.Level)
}

func (suite *PluginTestSuite) createTree() {
	electronics := Taxon{
		Name: "Electronics",
	}

	television := Taxon{
		Name:   "Television",
		Parent: &electronics,
	}
	gameConsoles := Taxon{
		Name:   "Game Consoles",
		Parent: &electronics,
	}
	portableElectronics := Taxon{
		Name:   "Portable Electronics",
		Parent: &electronics,
	}

	tube := Taxon{
		Name:   "Tube",
		Parent: &television,
	}
	lcd := Taxon{
		Name:   "LCD",
		Parent: &television,
	}
	plasma := Taxon{
		Name:   "Plasma",
		Parent: &television,
	}

	mp3 := Taxon{
		Name:   "MP3",
		Parent: &portableElectronics,
	}

	cdPlayer := Taxon{
		Name:   "CD Player",
		Parent: &portableElectronics,
	}

	radio := Taxon{
		Name:   "Radio",
		Parent: &portableElectronics,
	}

	flash := Taxon{
		Name:   "Flash",
		Parent: &mp3,
	}

	suite.db.Save(&television)
	suite.db.Save(&gameConsoles)
	suite.db.Save(&tube)
	suite.db.Save(&lcd)
	suite.db.Save(&plasma)
	suite.db.Save(&flash)
	suite.db.Save(&cdPlayer)
	suite.db.Save(&radio)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 11)
	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.Lft)
			assert.Equal(suite.T(), 22, taxon.Rgt)
			assert.Equal(suite.T(), 0, taxon.Level)
			count++
			break
		case "Television":
			assert.Equal(suite.T(), 2, taxon.Lft)
			assert.Equal(suite.T(), 9, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "Tube":
			assert.Equal(suite.T(), 3, taxon.Lft)
			assert.Equal(suite.T(), 4, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
		case "LCD":
			assert.Equal(suite.T(), 5, taxon.Lft)
			assert.Equal(suite.T(), 6, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Plasma":
			assert.Equal(suite.T(), 7, taxon.Lft)
			assert.Equal(suite.T(), 8, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 10, taxon.Lft)
			assert.Equal(suite.T(), 11, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 12, taxon.Lft)
			assert.Equal(suite.T(), 21, taxon.Rgt)
			assert.Equal(suite.T(), 1, taxon.Level)
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 13, taxon.Lft)
			assert.Equal(suite.T(), 16, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 17, taxon.Lft)
			assert.Equal(suite.T(), 18, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 19, taxon.Lft)
			assert.Equal(suite.T(), 20, taxon.Rgt)
			assert.Equal(suite.T(), 2, taxon.Level)
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 14, taxon.Lft)
			assert.Equal(suite.T(), 15, taxon.Rgt)
			assert.Equal(suite.T(), 3, taxon.Level)
			count++
			break
		}
	}

	assert.Equal(suite.T(), len(taxons), count)
}

func (suite *PluginTestSuite) TestGetTreeLeft() {
	t := &Taxon{
		Lft: 41,
	}

	assert.Equal(suite.T(), 41, nested.GetTreeLeft(t))
}

func (suite *PluginTestSuite) TestGetTreeRight() {
	t := &Taxon{
		Rgt: 41,
	}

	assert.Equal(suite.T(), 41, nested.GetTreeRight(t))
}

func (suite *PluginTestSuite) TestGetTreeLevel() {
	t := &Taxon{
		Level: 41,
	}

	assert.Equal(suite.T(), 41, nested.GetTreeLevel(t))
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}
