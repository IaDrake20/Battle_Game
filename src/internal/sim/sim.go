package sim

import (
        "time"

        "battle/src/internal/entity"
)

const TickRate = 20 // ticks per second

type World struct {
        Tick     uint64
        Entities map[string]*entity.Entity
}

func NewWorld() *World {
        return &World{Entities: make(map[string]*entity.Entity)}
}

func (w *World) Step(dt float64) {
        w.Tick++
        for _, e := range w.Entities {
                e.Update(dt)
        }
        // TODO: collision detection, combat resolution, requisition accrual
}

// Run drives the world at a fixed timestep, invoking onTick after each step.
func Run(world *World, onTick func(*World)) {
        ticker := time.NewTicker(time.Second / TickRate)
        defer ticker.Stop()
        dt := 1.0 / float64(TickRate)
        for range ticker.C {
                world.Step(dt)
                if onTick != nil {
                        onTick(world)
                }
        }
}
