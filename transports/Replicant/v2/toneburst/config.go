package toneburst

type Config interface {
	Construct() (ToneBurst, error)
}
