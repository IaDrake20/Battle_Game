package entity

type Kind string

const (
        KindShip      Kind = "ship"
        KindPlane     Kind = "plane"
        KindSubmarine Kind = "submarine"
        KindTroop     Kind = "troop"
)

type Vec3 struct {
        X, Y, Z float64
}

type Entity struct {
        ID      string
        Kind    Kind
        Faction string
        Pos     Vec3
        Heading float64
        Speed   float64
        Health  float64
        // TODO: status effects, cargo, orders/pathfinding target
}

func (e *Entity) Update(dt float64) {
        // TODO: apply pathfinding/movement, ballistics, status effect ticks
}
