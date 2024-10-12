package berrt

import (
	"fmt"
)

type DiagramType int
type PlantUMLType string

const (
	SeparatorPlantUMLType    PlantUMLType = "== %s =="
	RequestPlantUMLType      PlantUMLType = "%s -> %s : %s"
	ResponsePlantUMLType     PlantUMLType = "%s --> %s : %s"
	DelayPlantUMLType        PlantUMLType = "%s ...%s..."
	ParticipantPlantUMLType  PlantUMLType = "participant %s"
	ActivationPlantUMLType   PlantUMLType = "activate %s"
	DeactivationPlantUMLType PlantUMLType = "deactivate %s"
	DestroyPlantUMLType      PlantUMLType = "destroy %s"
	MapPlantUMLType          PlantUMLType = "map %s {\n\t<%s>\n}"
	MapTaskPlantUMLType      PlantUMLType = "task %s {\n\t%s => %s\n}"
)

func NewPlantUMLType(t string, values ...any) PlantUMLType {
	return PlantUMLType(fmt.Sprintf(t, values...))
}

func (p PlantUMLType) String() string {
	return string(p)
}
