package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// Sliders enumeration
const (
	ballsSlider = iota
	temperatureSlider
	widthSlider
	heightSlider
	gapSlider
	bricksSlider
	massSlider
	frictionSlider
	numberOfSliders
)

var sliderLabels = map[int]string{
	ballsSlider:       "balls",
	temperatureSlider: "temperature",
	widthSlider:       "width",
	heightSlider:      "height",
	gapSlider:         "gap",
	bricksSlider:      "bricks",
	massSlider:        "mass",
	frictionSlider:    "friction",
}

const (
	screenWidth  = 1900
	screenHeight = 1340

	ballRadius = 4
	ballMass   = 1.0

	epsilon = 0.1
)

// Ball is a single microscopic particle
type Ball struct {
	x, y    float64
	vx, vy  float64
	r       float64
	mass    float64
	clr     color.Color
	lastHit int
}

// Box is the particle that moves as a result of Brownian motion
type Box struct {
	x, y   float64
	vx, vy float64
	w, h   float64
	mass   float64
	clr    color.Color
}

type Slider struct {
	value      float64
	ival       int
	min, max   float64
	isInteger  bool
	x, y, w, h int
	label      string
}

type Game struct {
	balls []Ball
	boxes []Box

	ballsVisible bool
}

var (
	sliders []Slider
)

func (s *Slider) IntValue() int {
	if s.isInteger {
		return s.ival
	}
	return int(s.value)
}

func (s *Slider) Value() float64 {
	if s.isInteger {
		return float64(s.ival)
	}
	return s.value
}

func (s *Slider) SetValue(val float64) {
	if s.isInteger {
		s.ival = int(val)
		s.value = float64(s.ival)
		fmt.Println(s.ival)
	} else {
		s.value = val
		s.ival = int(val)
		fmt.Println(val)
	}
}

func (s *Slider) Click(x, y int) {
	const margin = 5
	const paddingLeft = 5
	const paddingRight = 3
	x -= s.x
	y -= s.y
	if x < -margin || x >= s.w+margin || y < 0 || y >= s.h {
		return
	}

	val := float64(x-paddingLeft)*(s.max-s.min)/float64(s.w-paddingLeft-paddingRight) + s.min
	if val < s.min {
		val = s.min
	}
	if val > s.max {
		val = s.max
	}
	s.SetValue(val)
}

func (s *Slider) Draw(screen *ebiten.Image) {
	const gap = 2
	const gap2 = 4
	const margin = 2

	vector.DrawFilledRect(screen, float32(s.x-margin), float32(s.y-margin),
		float32(s.w+2*margin), float32(s.h+2*margin), color.RGBA{40, 180, 255, 255}, false)

	vector.DrawFilledRect(screen, float32(s.x), float32(s.y), float32(s.w), float32(s.h),
		color.RGBA{200, 200, 200, 255}, false)

	vector.DrawFilledRect(screen, float32(s.x+gap), float32(s.y+gap), float32(s.w-2*gap), float32(s.h-2*gap),
		color.RGBA{30, 5, 0, 255}, false)

	w := float64(s.w-gap-gap2) * (s.Value() - s.min) / (s.max - s.min)
	vector.DrawFilledRect(screen, float32(s.x+gap2), float32(s.y+gap2), float32(w), float32(s.h-gap-gap2),
		color.RGBA{190, 25, 0, 255}, false)

	var str string
	if s.isInteger {
		str = fmt.Sprintf("%d", s.IntValue())
	} else {
		str = fmt.Sprintf("%.2f", s.Value())
	}
	str = fmt.Sprintf("%s = %s", s.label, str)
	text.Draw(screen, str, basicfont.Face7x13, s.x+12, s.y+(s.h+8)/2, color.White)
}

func drawSliders(screen *ebiten.Image, sliders []Slider) {
	for i := range sliders {
		sliders[i].Draw(screen)
	}
}

func NewGame() *Game {
	g := &Game{}
	g.Restart()
	return g
}

func (g *Game) Restart() {
	numBalls := sliders[ballsSlider].IntValue()
	maxTemperature := sliders[temperatureSlider].Value()
	numBricks := sliders[bricksSlider].IntValue()
	boxWidth := sliders[widthSlider].Value()
	boxHeight := sliders[heightSlider].Value()
	gap := sliders[gapSlider].Value()
	boxMass := sliders[massSlider].Value()

	balls := make([]Ball, numBalls)
	for i := range balls {
		x := float64(ballRadius + rand.Intn(screenWidth-2*ballRadius))
		y := float64(ballRadius + rand.Intn(screenHeight-2*ballRadius))
		vx := rand.Float64()*2*maxTemperature - maxTemperature
		vy := rand.Float64()*2*maxTemperature - maxTemperature

		balls[i] = Ball{
			x:       x,
			y:       y,
			vx:      vx,
			vy:      vy,
			r:       ballRadius,
			mass:    ballMass,
			clr:     randomColor(),
			lastHit: -1,
		}
	}

	boxes := make([]Box, numBricks)
	for i := range boxes {
		x := (float64(screenWidth)-float64(numBricks)*boxWidth-float64(numBricks-1)*gap)/2 + float64(i)*(boxWidth+gap)
		y := float64(screenHeight-boxHeight) / 2

		vx, vy := 0.0, 0.0

		boxes[i] = Box{
			x:    x,
			y:    y,
			vx:   vx,
			vy:   vy,
			w:    boxWidth,
			h:    boxHeight,
			mass: boxMass,
			clr:  color.RGBA{R: 200, G: 30, B: 0, A: 255},
		}
	}

	g.balls = balls
	g.boxes = boxes
	g.ballsVisible = true
}

func initSliders() {
	sliders = make([]Slider, numberOfSliders)
	setSlider(ballsSlider,        0,  5000, 350,  true)
	setSlider(temperatureSlider,  0,    15,   5, false)
	setSlider(widthSlider,       20,   600, 250, false)
	setSlider(heightSlider,      20,  1200, 500, false)
	setSlider(gapSlider,          0,   300,  40, false)
	setSlider(bricksSlider,       0,    10,   1,  true)
	setSlider(massSlider,       0.1, 10000, 600, false)
	setSlider(frictionSlider,     0,     1,   0, false)
}

func setSlider(index int, min, max, value float64, isInteger bool) {
	const startX = 20
	const startY = 20
	const gapY = 15
	const width = 180
	const height = 40

	y := startY
	if index != 0 {
		y = sliders[index-1].y + sliders[index-1].h + gapY
	}

	sliders[index] = Slider{
		x:         startX,
		y:         y,
		w:         width,
		h:         height,
		min:       min,
		max:       max,
		value:     value,
		ival:      int(value),
		isInteger: isInteger,
		label:     sliderLabels[index],
	}
}

func (g *Game) Update() error {
	// 1. Move balls
	for i := range g.balls {
		g.balls[i].x += g.balls[i].vx
		g.balls[i].y += g.balls[i].vy
		// Collide with walls
		checkBallWallCollision(&g.balls[i])
	}
  // 1.1. Move boxes
	for i := range g.boxes {
		g.boxes[i].x += g.boxes[i].vx
		g.boxes[i].y += g.boxes[i].vy
    // Collide with walls
		checkBoxWallCollision(&g.boxes[i])
	}

	// 2. Ball-ball collisions
	for i := 0; i < len(g.balls); i++ {
		for j := i + 1; j < len(g.balls); j++ {
			collideBalls(&g.balls[i], &g.balls[j])
		}
	}

	// 3. Ball-box collisions
	for i := range g.balls {
		for j := range g.boxes {
			collideBallBox(&g.balls[i], &g.boxes[j], j)
		}
	}

	// 4. Box-box collisions
	for i := 0; i < len(g.boxes); i++ {
		for j := i + 1; j < len(g.boxes); j++ {
			collideBoxes(&g.boxes[i], &g.boxes[j])
		}
	}

	// Box friction
	for i := range g.boxes {
		frictBoxes(g, &g.boxes[i])
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		for i := range sliders {
			sliders[i].Click(x, y)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.ballsVisible = !g.ballsVisible
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.Restart()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	if g.ballsVisible {
		for _, b := range g.balls {
			drawCircle(screen, b.x, b.y, b.r, b.clr)
		}
	}

	for _, box := range g.boxes {
		vector.DrawFilledRect(screen, float32(box.x), float32(box.y), float32(box.w), float32(box.h), box.clr, false)
	}

	drawSliders(screen, sliders)

	n := len(sliders) - 1
	x := sliders[n].x
	y := sliders[n].y + sliders[n].h + 30
	text.Draw(screen, "Enter - apply & restart", basicfont.Face7x13, x, y, color.White)
	text.Draw(screen, "Space - show/hide atoms", basicfont.Face7x13, x, y+17, color.White)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func checkBallWallCollision(b *Ball) {
	if b.x-b.r < 0 {
		b.x = b.r
		b.vx = -b.vx
	} else if b.x+b.r > screenWidth {
		b.x = screenWidth - b.r
		b.vx = -b.vx
	}

	if b.y-b.r < 0 {
		b.y = b.r
		b.vy = -b.vy
	} else if b.y+b.r > screenHeight {
		b.y = screenHeight - b.r
		b.vy = -b.vy
	}
}

func checkBoxWallCollision(box *Box) {
	if box.x < 0 {
		box.x = 0
		box.vx = -box.vx
	} else if box.x+box.w > screenWidth {
		box.x = screenWidth - box.w
		box.vx = -box.vx
	}

	if box.y < 0 {
		box.y = 0
		box.vy = -box.vy
	} else if box.y+box.h > screenHeight {
		box.y = screenHeight - box.h
		box.vy = -box.vy
	}
}

func collideBalls(b1, b2 *Ball) {
	dx := b2.x - b1.x
	dy := b2.y - b1.y
	dist := math.Hypot(dx, dy)

	if dist >= b1.r+b2.r || dist == 0 {
		return
	}

	nx := dx / dist
	ny := dy / dist

	v1 := b1.vx*nx + b1.vy*ny
	v2 := b2.vx*nx + b2.vy*ny

	m1 := b1.mass
	m2 := b2.mass

	v1p := (v1*(m1-m2) + 2*m2*v2) / (m1 + m2)
	v2p := (v2*(m2-m1) + 2*m1*v1) / (m1 + m2)

	b1.vx += (v1p - v1) * nx
	b1.vy += (v1p - v1) * ny
	b2.vx += (v2p - v2) * nx
	b2.vy += (v2p - v2) * ny

	overlap := (b1.r + b2.r) - dist
	b1.x -= overlap / 2 * nx
	b1.y -= overlap / 2 * ny
	b2.x += overlap / 2 * nx
	b2.y += overlap / 2 * ny
}

func collideBallBox(b *Ball, box *Box, boxIndex int) {
	if b.lastHit == boxIndex {
		b.lastHit = -1
		return
	}

	closestX := math.Max(box.x, math.Min(b.x, box.x+box.w))
	closestY := math.Max(box.y, math.Min(b.y, box.y+box.h))

	dx := closestX - b.x
	dy := closestY - b.y
	dist2 := dx*dx + dy*dy

	if dist2 < epsilon {
		return
	}

	r2 := b.r * b.r
	if dist2 > r2 {
		return
	}

	if math.IsNaN(b.x) {
		log.Fatal("Fatal error: b.x = NaN")
	}

	dist := math.Sqrt(dist2)

	m1 := b.mass
	m2 := box.mass

	nx, ny := dx/dist, dy/dist

	vBall := b.vx*nx + b.vy*ny
	vBox := box.vx*nx + box.vy*ny

	vBallP := (vBall*(m1-m2) + 2*m2*vBox) / (m1 + m2)
	vBoxP := (vBox*(m2-m1) + 2*m1*vBall) / (m1 + m2)

	bvx, bvy := b.vx, b.vy

	b.vx += (vBallP - vBall) * nx
	b.vy += (vBallP - vBall) * ny

	box.vx += (vBoxP - vBox) * nx
	box.vy += (vBoxP - vBox) * ny

	b.x -= bvx
	b.y -= bvy

	b.lastHit = boxIndex
}

func collideBoxes(b1, b2 *Box) {
	if b1.x+b1.w < b2.x || b2.x+b2.w < b1.x ||
		b1.y+b1.h < b2.y || b2.y+b2.h < b1.y {
		return
	}

	overlapX1 := (b1.x + b1.w) - b2.x
	overlapX2 := (b2.x + b2.w) - b1.x
	overlapY1 := (b1.y + b1.h) - b2.y
	overlapY2 := (b2.y + b2.h) - b1.y

	dx := math.Min(overlapX1, overlapX2)
	dy := math.Min(overlapY1, overlapY2)

	if dx < dy {
		if b1.x < b2.x {
			b1.x -= dx / 2
			b2.x += dx / 2
		} else {
			b1.x += dx / 2
			b2.x -= dx / 2
		}
		b1.vx = -b1.vx
		b2.vx = -b2.vx
	} else {
		if b1.y < b2.y {
			b1.y -= dy / 2
			b2.y += dy / 2
		} else {
			b1.y += dy / 2
			b2.y -= dy / 2
		}
		b1.vy = -b1.vy
		b2.vy = -b2.vy
	}
}

func drawCircle(screen *ebiten.Image, cx, cy float64, r float64, clr color.Color) {
	segments := 36
	angleStep := 2 * math.Pi / float64(segments)
	var x0, y0 float64
	x0 = cx + r*math.Cos(0)
	y0 = cy + r*math.Sin(0)
	for i := 1; i <= segments; i++ {
		angle := float64(i) * angleStep
		x1 := cx + r*math.Cos(angle)
		y1 := cy + r*math.Sin(angle)
		ebitenutil.DrawLine(screen, x0, y0, x1, y1, clr)
		x0, y0 = x1, y1
	}
}

func frictBoxes(g *Game, b *Box) {
	f := 1 - sliders[frictionSlider].Value()
	b.vx *= f
	b.vy *= f
}

func randomColor() color.Color {
	return color.RGBA{
		R: 0,
		G: uint8(rand.Intn(150) + 50),
		B: 255,
		A: 255,
	}
}

func main() {
	initSliders()
	game := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Balls")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
