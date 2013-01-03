package main

import "image"
import "image/color"
import _ "image/gif"
import _ "image/jpeg"
import "image/png"
import "math/rand"
import "time"
import "os"
import "fmt"
import "sort"

const rectCount = 600
const rectMaxSize = 30
const entityCount = 200
const entityKeep = 20
const mutationFactor = 0.05

type ShadedRectangle struct {
	color  color.Gray
	bounds image.Rectangle
}

type Entity struct {
	width, height int
	bgColor       color.Gray
	rects         []ShadedRectangle
}

func randomEntity(width, height, rectCount int) *Entity {
	rects := make([]ShadedRectangle, rectCount)

	for i := 0; i < len(rects); i++ {
		color := color.Gray{uint8(rand.Intn(256))}
		x1, x2 := rand.Intn(width), rand.Intn(width)
		y1, y2 := rand.Intn(height), rand.Intn(height)
		rect := image.Rect(x1, y1, x2, y2).Canon()

		/* enforce max size */
		if rect.Max.X-rect.Min.X > rectMaxSize {
			overage := rect.Max.X - rect.Min.X - rectMaxSize
			rect.Max.X -= overage / 2
			rect.Min.X += overage / 2
		}

		if rect.Max.Y-rect.Min.Y > rectMaxSize {
			overage := rect.Max.Y - rect.Min.Y - rectMaxSize
			rect.Max.Y -= overage / 2
			rect.Min.Y += overage / 2
		}

		rects[i] = ShadedRectangle{color, rect}
	}

	bgColor := color.Gray{uint8(rand.Intn(256))}

	return &Entity{width, height, bgColor, rects}
}

func (entity *Entity) render() *image.Gray {
	image := image.NewGray(image.Rect(0, 0, entity.width, entity.height))

	for x := 0; x < entity.width; x++ {
		for y := 0; y < entity.height; y++ {
			image.SetGray(x, y, entity.bgColor)
		}
	}

	for i := 0; i < len(entity.rects); i++ {
		rect := &entity.rects[i]

		for x := rect.bounds.Min.X; x < rect.bounds.Max.X; x++ {
			for y := rect.bounds.Min.Y; y < rect.bounds.Max.Y; y++ {
				image.SetGray(x, y, rect.color)
			}
		}
	}

	return image
}

func (entity *Entity) deviation(model *image.Gray) float64 {
	deviation := 0.0
	actual := entity.render()

	for x := 0; x < entity.width; x++ {
		for y := 0; y < entity.height; y++ {
			actualColor := float64(actual.At(x, y).(color.Gray).Y)
			modelColor := float64(model.At(x, y).(color.Gray).Y)
			dev := actualColor - modelColor
			deviation += dev * dev
		}
	}

	return deviation
}

func (entity *Entity) reproduce(other *Entity) *Entity {
	newRects := make([]ShadedRectangle, len(entity.rects))
	for i := 0; i < len(entity.rects); i++ {
		if rand.Intn(2) == 0 {
			newRects[i] = entity.rects[i]
		} else {
			newRects[i] = other.rects[i]
		}

		newRects[i].color.Y += uint8(2 * mutationFactor * (rand.Float64() - 0.5) * 256)
		newRects[i].bounds.Min.X += int(2 * mutationFactor * (rand.Float64() - 0.5) * float64(entity.width))
		newRects[i].bounds.Max.X += int(2 * mutationFactor * (rand.Float64() - 0.5) * float64(entity.width))
		newRects[i].bounds.Min.Y += int(2 * mutationFactor * (rand.Float64() - 0.5) * float64(entity.height))
		newRects[i].bounds.Max.Y += int(2 * mutationFactor * (rand.Float64() - 0.5) * float64(entity.height))

		newRects[i].bounds.Min.X = clip(newRects[i].bounds.Min.X, 0, entity.width)
		newRects[i].bounds.Max.X = clip(newRects[i].bounds.Max.X, 0, entity.width)
		newRects[i].bounds.Min.Y = clip(newRects[i].bounds.Min.Y, 0, entity.height)
		newRects[i].bounds.Max.Y = clip(newRects[i].bounds.Max.Y, 0, entity.height)

		newRects[i].bounds = newRects[i].bounds.Canon()

		/* enforce max size */
		if newRects[i].bounds.Max.X-newRects[i].bounds.Min.X > rectMaxSize {
			overage := newRects[i].bounds.Max.X - newRects[i].bounds.Min.X - rectMaxSize
			newRects[i].bounds.Max.X -= overage / 2
			newRects[i].bounds.Min.X += overage / 2
		}

		if newRects[i].bounds.Max.Y-newRects[i].bounds.Min.Y > rectMaxSize {
			overage := newRects[i].bounds.Max.Y - newRects[i].bounds.Min.Y - rectMaxSize
			newRects[i].bounds.Max.Y -= overage / 2
			newRects[i].bounds.Min.Y += overage / 2
		}
	}

	var newBgColor color.Gray
	if rand.Intn(2) == 0 {
		newBgColor = entity.bgColor
	} else {
		newBgColor = other.bgColor
	}
	newBgColor.Y += uint8(2 * mutationFactor * (rand.Float64() - 0.5) * 256)

	return &Entity{entity.width, entity.height, newBgColor, newRects}
}

func clip(value, min, max int) int {
	switch {
	case value < min:
		return min
	case value > max:
		return max
	}

	return value
}

type deviancyInfo struct {
	entity   *Entity
	deviancy float64
}

type deviancyInfos []deviancyInfo

func (di deviancyInfos) Len() int           { return len(di) }
func (di deviancyInfos) Less(i, j int) bool { return di[i].deviancy < di[j].deviancy }
func (di deviancyInfos) Swap(i, j int)      { di[i], di[j] = di[j], di[i] }

func step(entities []*Entity, model *image.Gray) float64 {
	/* determine deviancy */
	deviancies := make([]deviancyInfo, len(entities))
	deviancyCh := make(chan deviancyInfo)

	for i := 0; i < len(entities); i++ {
		go func(entity *Entity) {
			deviancyCh <- deviancyInfo{entity, entity.deviation(model)}
		}(entities[i])
	}

	for i := 0; i < len(entities); i++ {
		deviancies[i] = <-deviancyCh
	}

	/* sort */
	sort.Sort(deviancyInfos(deviancies))

	/* reproduce */
	for i := 0; i < entityKeep; i++ {
		entities[i] = deviancies[i].entity
	}

	entityCh := make(chan *Entity)
	for i := entityKeep; i < len(entities); i++ {
		go func(entity, other *Entity) {
			entityCh <- entity.reproduce(other)
		}(entities[(i-entityKeep)%entityKeep], entities[((i-entityKeep)/entityKeep)%entityKeep])
	}

	for i := entityKeep; i < len(entities); i++ {
		entities[i] = <-entityCh
	}

	return deviancies[0].deviancy
}

func main() {
	rand.Seed(time.Now().UnixNano())

	/* read model image */
	if len(os.Args) < 2 {
		fmt.Println("error: need an image file name")
		return
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}
	modelColor, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	/* convert model image to grayscale */
	bounds := modelColor.Bounds()
	model := image.NewGray(image.Rect(0, 0, bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y))
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			model.Set(x-bounds.Min.X, y-bounds.Min.Y, modelColor.At(x, y))
		}
	}
	modelColor = nil

	/* create inital conditions */
	fmt.Println("creating generation 0...")
	entities := make([]*Entity, entityCount)
	for i := 0; i < len(entities); i++ {
		entities[i] = randomEntity(model.Bounds().Max.X, model.Bounds().Max.Y, rectCount)
	}

	/* evolve */
	var best *Entity
	generation := 0
	for {
		generation++
		fmt.Printf("evolving generation %d...\n", generation)
		dev := step(entities, model)

		if best != entities[0] {
			best = entities[0]
			fmt.Printf(" ...new optimum found (deviancy = %f)\n", dev)

			/* once step is complete, the first element of entities is the most fit entity from the previous generation */
			file, err = os.Create(fmt.Sprintf("out/%09d.png", generation-1))
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return
			}
			png.Encode(file, entities[0].render())
			file.Close()
		}
	}
}
