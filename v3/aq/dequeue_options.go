package aq

type DequeMode int

const (
	Browse       DequeMode = 1
	Locked       DequeMode = 2
	Remove       DequeMode = 3
	RemoveNoData DequeMode = 4
)

type NavigationMode int

const (
	FirstMessage           NavigationMode = 1
	NextTransaction        NavigationMode = 2
	NextMessage            NavigationMode = 3
	FirstMessageMultiGroup NavigationMode = 4
	NextMessageMultiGroup  NavigationMode = 5
)

type DequeueOptions struct {
	Mode           DequeMode
	Navigation     NavigationMode
	Visibility     VisibilityMode
	Delivery       DeliveryMode
	Wait           int
	Consumer       string
	Correlation    string
	Condition      string
	Transformation string
}

func DefaultDequeueOptions() *DequeueOptions {
	return &DequeueOptions{
		Mode:       Remove,
		Navigation: NextMessage,
		Visibility: VisibilityOnCommit,
		Delivery:   DeliveryModePersistent,
		Wait:       10,
	}
}
