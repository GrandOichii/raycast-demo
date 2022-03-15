package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	metaDataFile = "metadata.json"
)

// Map data
type MapData struct {
	Speed       float64             `json:"speed"`     // The speed of the player
	TurnSpeed   int                 `json:"turnspeed"` // The turn speed of the player (in degrees)
	SpawnCoords [2]float64          `json:"spawn"`     // The coords of the spawn point
	Tiles       map[string]TileData `json:"tiles"`     // The map tiles
}

// Tile data
type TileData struct {
	Passable bool    `json:"passable"` // Shows whether the player can pass through the tile (if the tile is a floor tile)
	Colors   []int16 `json:"colors"`   // The range of colors for the tile
}

// Loads the map data from the metadata file
func loadMD(dir string) (*MapData, error) {
	result := MapData{}
	data, err := os.ReadFile(dir)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &result)
	return &result, err
}

// Game map
type Map struct {
	md    *MapData   // Map data
	tiles [][]string // Tiles
	h, w  int        // dimensions
}

// Loads the map from a .map file (the directory also has to contain a metadata file)
func Load(dir string) (*Map, error) {
	data, err := os.ReadFile(dir)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	result := Map{}
	// load the tiles themselves
	result.tiles = make([][]string, 0, len(lines))
	for _, line := range lines {
		result.tiles = append(result.tiles, strings.Split(line, ""))
	}
	// get the metadata file
	homeDir := filepath.Dir(dir)
	mdf := fmt.Sprintf("%s/%s", homeDir, metaDataFile)
	md, err := loadMD(mdf)
	if err != nil {
		return nil, err
	}
	result.md = md
	result.h = len(lines)
	result.w = len(lines[0])
	return &result, nil
}
