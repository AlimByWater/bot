package entity

const (
	ChatTypeSuperGroup = "supergroup"
)

type InitWebApp struct {
	InitData        string `json:"init_data"`
	AnimationToShow string `json:"animation_to_show"`
}
