package main

import (
	"fmt"
	"math"
	"os"

	nc "github.com/rthornton128/goncurses"
)

const (
	borderColor = nc.C_MAGENTA

	coneView = 89 // the view cone of the player
	maxDist  = 20

	ceilingChar = " "
	wallChar    = "#"
	floorChar   = "."
)

// color
var (
	// color pair counter
	cpi int16 = 0

	// border color pair
	borderCP nc.Char
)

var (
	// Dimensions of the window
	wHeight, wWidth int

	// player pos
	pY, pX  float64
	pA      int        = 0
	running bool       // True if the window is running
	win     *nc.Window // The window
	m       *Map       // Game map

	tileColorMap = map[string][]nc.Char{} // Color pair ranges for the maps

	keyFunc = map[nc.Key]func() (newY float64, newX float64, err error){ // Key to func map
		nc.KEY_ESC: func() (newY float64, newX float64, err error) {
			running = false
			return pY, pX, nil
		},
		nc.KEY_UP: func() (newY float64, newX float64, err error) {
			a := pRadA()
			s := m.md.Speed
			return pY + s*math.Sin(a), pX + s*math.Cos(a), nil
		},
		nc.KEY_DOWN: func() (newY float64, newX float64, err error) {
			a := pRadA()
			s := m.md.Speed
			return pY - s*math.Sin(a), pX - s*math.Cos(a), nil
		},
		nc.KEY_LEFT: func() (newY float64, newX float64, err error) {
			pA = (pA - m.md.TurnSpeed) % 360
			return pY, pX, nil
		},
		nc.KEY_RIGHT: func() (newY float64, newX float64, err error) {
			pA = (pA + m.md.TurnSpeed) % 360
			return pY, pX, nil
		},
		'<': func() (newY float64, newX float64, err error) {
			a := pRadA()
			s := m.md.Speed
			return pY - s*math.Cos(a), pX + s*math.Sin(a), nil
		},
		'>': func() (newY float64, newX float64, err error) {
			a := pRadA()
			s := m.md.Speed
			return pY + s*math.Cos(a), pX - s*math.Sin(a), nil
		},
	}
)

// Returns the player angle in rads
func pRadA() float64 {
	return float64(pA) * math.Pi / 180
}

// Ends the window and panics if the error is not nil
func checkErr(err error) {
	if err != nil {
		nc.End()
		panic(err)
	}
}

// Entry point
func main() {
	var err error
	if len(os.Args) != 2 {
		fmt.Printf("Please specify the path to the map file (the directory also has to contain the %s file)\n", metaDataFile)
		return
	}
	// load the map
	m, err = Load(os.Args[1])
	// m, err = Load("maps/map1/map.map")
	checkErr(err)
	err = setup()
	checkErr(err)
	err = start()
	checkErr(err)
	nc.End()
}

// Draws a border around a window
func drawBorder(w *nc.Window, a ...nc.Char) error {
	for _, at := range a {
		w.AttrOn(at)
		defer w.AttrOff(at)
	}
	err := win.Border(nc.ACS_VLINE, nc.ACS_VLINE, nc.ACS_HLINE, nc.ACS_HLINE, nc.ACS_ULCORNER, nc.ACS_URCORNER, nc.ACS_LLCORNER, nc.ACS_LRCORNER)
	return err
}

// Setups the window and the colors
func setup() error {
	var err error
	// init the window
	win, err = nc.Init()
	checkErr(err)
	nc.Cursor(0)
	nc.SetEscDelay(0)
	nc.Echo(false)
	win.Keypad(true)
	wHeight, wWidth = win.MaxYX()
	// init the colors
	err = nc.StartColor()
	if err != nil {
		return err
	}
	err = nc.UseDefaultColors()
	if err != nil {
		return err
	}
	cpi++
	err = nc.InitPair(cpi, borderColor, -1)
	if err != nil {
		return err
	}
	borderCP = nc.ColorPair(cpi)
	// init tile color ranges
	for t, td := range m.md.Tiles {
		colors := []nc.Char{}
		for _, c := range td.Colors {
			cpi++
			err = nc.InitPair(cpi, -1, c)
			if err != nil {
				return err
			}
			colors = append(colors, nc.ColorPair(cpi))
		}
		tileColorMap[t] = colors
	}
	return nil
}

// Starts the engine
func start() error {
	var err error
	err = initialDraw()
	if err != nil {
		return err
	}
	pY = m.md.SpawnCoords[0]
	pX = m.md.SpawnCoords[1]
	running = true
	for running {
		err = draw()
		if err != nil {
			return err
		}
		err = handleInput()
		if err != nil {
			return err
		}
	}
	return nil
}

// The initial draw of the window
func initialDraw() error {
	err := drawBorder(win, borderCP)
	return err
}

// The draw function
func draw() error {
	repeat := (wWidth - 2) / coneView
	hcv := coneView / 2
	for i := -hcv; i < hcv; i++ {
		ceilc, wallc, floorc, tile, err := castRay(i)
		if err != nil {
			return err
		}
		j := 0
		// draw the ceiling
		for ci := 0; ci < ceilc; ci++ {
			for ia := 0; ia < repeat; ia++ {
				win.MovePrint(1+j, 1+(i+hcv)*repeat+ia, ceilingChar)
			}
			j++
		}
		// draw the wall
		colors, has := tileColorMap[tile]
		if !has {
			if tile == "" {
				colors = []nc.Char{nc.Char(0)}
			} else {
				return fmt.Errorf("unrecognizable char: %s", tile)
			}
		}
		// vWidth-2 - len(colors)-1
		// wallc - ?
		ci := (len(colors) - 1) * wallc / (wHeight - 2)
		colorPair := colors[ci]
		win.AttrOn(colorPair)
		for wi := 0; wi < wallc; wi++ {
			for ia := 0; ia < repeat; ia++ {
				win.MoveAddChar(1+j, 1+(i+hcv)*repeat+ia, ' ')
			}
			j++
		}
		win.AttrOff(colorPair)
		// draw the floor
		for fi := 0; fi < floorc; fi++ {
			for ia := 0; ia < repeat; ia++ {
				win.MovePrint(1+j, 1+(i+hcv)*repeat+ia, floorChar)
			}
			j++
		}
	}
	return nil
}

// Casts a ray
func castRay(rc int) (int, int, int, string, error) {
	deg := (pA + rc) % 360
	rad := float64(deg) * math.Pi / 180
	i := 1.0
	var fy, fx float64 = 0, 0
	var y, x int
	var tc string
	for {
		fy = pY + math.Sin(rad)*i
		fx = pX + math.Cos(rad)*i
		y = int(fy)
		x = int(fx)
		if y >= m.h || x >= m.w || y < 0 || x < 0 {
			c, w, f := toColumn(0)
			return c, w, f, "", nil
		}
		tc = m.tiles[y][x]
		data, has := m.md.Tiles[tc]
		if !has {
			return 0, 0, 0, "", fmt.Errorf("unrecognizable char: %s", tc)
		}
		if !data.Passable {
			break
		}
		i += 0.1
		if i > maxDist {
			c, w, f := toColumn(0)
			return c, w, f, "", nil
		}
	}
	dist := math.Sqrt(math.Pow(pY-fy, 2) + math.Pow(pX-fx, 2))
	c, w, f := toColumn(dist)
	return c, w, f, m.tiles[y][x], nil
}

// Converts distance to a column
func toColumn(dist float64) (int, int, int) {
	if dist == 0 {
		half := (wHeight - 2) / 2
		return half, 0, half
	}
	fwh := float64(wHeight - 2)
	ceil := 0
	if int(dist) != 1 && int(dist) != 0 {
		ceil = int(math.Abs(fwh/2.0 - fwh/dist))
	}
	floor := ceil
	wall := wHeight - 2 - 2*ceil
	return ceil, wall, floor
}

// The handle input function
//
// Uses the keys from the keyFunc map
func handleInput() error {
	key := win.GetChar()
	f, has := keyFunc[key]
	if !has {
		return nil
	}
	newY, newX, err := f()
	if err != nil {
		return err
	}
	if newY < 0 || newX < 0 || newY >= float64(m.h) || newX >= float64(m.w) {
		return nil
	}
	tile := m.tiles[int(newY)][int(newX)]
	data, has := m.md.Tiles[tile]
	if !has {
		return fmt.Errorf("unrecognizable char: %s", tile)
	}
	// correct the values
	if !data.Passable {
		return nil
	}
	pY, pX = newY, newX
	return nil
}
