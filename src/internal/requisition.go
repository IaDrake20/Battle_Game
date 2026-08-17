package requisition

type PointKind string

const (
        Depot   PointKind = "depot"
        Fort    PointKind = "fort"
        SeaZone PointKind = "sea_zone" // one-time capture
)

type CapturePoint struct {
        ID       string
        Kind     PointKind
        Owner    string // faction ID; empty if neutral
        Rate     float64
        OneTime  bool
        Captured bool // for OneTime points
}

// Ledger tracks each faction's requisition balance and the cap on that
// balance, which is derived from the number and type of points held.
type Ledger struct {
        Balances map[string]float64
        Caps     map[string]float64
}

func NewLedger() *Ledger {
        return &Ledger{
                Balances: make(map[string]float64),
                Caps:     make(map[string]float64),
        }
}

func (l *Ledger) Accrue(points []*CapturePoint, dt float64) {
        for _, p := range points {
                if p.Owner == "" || (p.OneTime && p.Captured) {
                        continue
                }
                if p.OneTime {
                        l.Balances[p.Owner] += p.Rate
                        p.Captured = true
                        continue
                }
                bal := l.Balances[p.Owner] + p.Rate*dt
                if cap, ok := l.Caps[p.Owner]; ok && bal > cap {
                        bal = cap
                }
                l.Balances[p.Owner] = bal
        }
}
