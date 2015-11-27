package team

import (
	"testing"

	"fmt"
)

func TestGetTeamList(t *testing.T) {
	ti := &TeamInfo{
		Uid:      100000,
		TeamName: "abc_1_test",
	}
	if err := CreateTeam(ti); err != nil {
		fmt.Println(err)
		return
	}
	ti.TeamId = uint64(0)
	ti.TeamName = "abc_2_test"
	if err := CreateTeam(ti); err != nil {
		fmt.Println(err)
		return
	}

	ti.TeamId = uint64(0)
	ti.TeamName = "bac_3_test"
	if err := CreateTeam(ti); err != nil {
		fmt.Println(err)
		return
	}

	sel := map[string]interface{}{
		"teamname": map[string]interface{}{"$regex": "test", "$options": "$i"},
	}

	tl, err := GetTeamList(sel, 0, 30)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tl)
}
