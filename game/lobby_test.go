package game

import (
	"sync"
	"testing"
)

func createLobbyWithDemoPlayers(playercount int) *Lobby {
	owner := &Player{}
	lobby := &Lobby{
		Owner:   owner,
		creator: owner,
		mutex:   &sync.Mutex{},
	}
	for i := 0; i < playercount; i++ {
		lobby.players = append(lobby.players, &Player{
			Connected: true,
		})
	}

	return lobby
}

func Test_CalculateVotesNeededToKick(t *testing.T) {
	t.Run("Check necessary kick vote amount for players", func(test *testing.T) {
		var expectedResults = map[int]int{
			//Kinda irrelevant since you can't kick yourself, but who cares.
			1:  2,
			2:  2,
			3:  2,
			4:  2,
			5:  3,
			6:  3,
			7:  4,
			8:  4,
			9:  5,
			10: 5,
		}

		for playerCount, expctedRequiredVotes := range expectedResults {
			lobby := createLobbyWithDemoPlayers(playerCount)
			result := calculateVotesNeededToKick(nil, lobby)
			if result != expctedRequiredVotes {
				t.Errorf("Error. Necessary vote amount was %d, but should've been %d", result, expctedRequiredVotes)
			}
		}
	})
}

func Test_RemoveAccents(t *testing.T) {
	t.Run("Check removing accented characters", func(test *testing.T) {
		var expectedResults = map[string]string{
			"é":           "e",
			"É":           "E",
			"à":           "a",
			"À":           "A",
			"ç":           "c",
			"Ç":           "C",
			"ö":           "oe",
			"Ö":           "OE",
			"œ":           "oe",
			"\n":          "\n",
			"\t":          "\t",
			"\r":          "\r",
			"":            "",
			"·":           "·",
			"?:!":         "?:!",
			"ac-ab":       "acab",
			"ac - _ab-- ": "acab",
		}

		for k, v := range expectedResults {
			result := simplifyText(k)
			if result != v {
				t.Errorf("Error. Char was %s, but should've been %s", result, v)
			}
		}
	})
}

func Test_simplifyText(t *testing.T) {
	//We only test the replacement we do ourselves. We won't test
	//the "sanitize", or furthermore our expectations of it for now.

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "dash",
			input: "-",
			want:  "",
		},
		{
			name:  "underscore",
			input: "_",
			want:  "",
		},
		{
			name:  "space",
			input: " ",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := simplifyText(tt.input); got != tt.want {
				t.Errorf("simplifyText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_recalculateRanks(t *testing.T) {
	lobby := &Lobby{
		mutex: &sync.Mutex{},
	}
	lobby.players = append(lobby.players, &Player{
		ID:        "a",
		Score:     1,
		Connected: true,
	})
	lobby.players = append(lobby.players, &Player{
		ID:        "b",
		Score:     1,
		Connected: true,
	})
	recalculateRanks(lobby)

	rankPlayerA := lobby.players[0].Rank
	rankPlayerB := lobby.players[1].Rank
	if rankPlayerA != 1 || rankPlayerB != 1 {
		t.Errorf("With equal score, ranks should be equal. (A: %d; B: %d)",
			rankPlayerA, rankPlayerB)
	}

	lobby.players = append(lobby.players, &Player{
		ID:        "c",
		Score:     0,
		Connected: true,
	})
	recalculateRanks(lobby)

	rankPlayerA = lobby.players[0].Rank
	rankPlayerB = lobby.players[1].Rank
	if rankPlayerA != 1 || rankPlayerB != 1 {
		t.Errorf("With equal score, ranks should be equal. (A: %d; B: %d)",
			rankPlayerA, rankPlayerB)
	}

	rankPlayerC := lobby.players[2].Rank
	if rankPlayerC != 2 {
		t.Errorf("new player should be rank 2, since the previous two players had the same rank. (C: %d)", rankPlayerC)
	}
}

func Test_calculateGuesserScore(t *testing.T) {
	// Test with a 5-character word, 115 seconds left out of 120 total
	lastScore := calculateGuesserScore(5, 115, 120)
	expectedScore := 10*5 + 100*115/120 // 50 + 95.83... = 145
	if lastScore != expectedScore {
		t.Errorf("Expected score %d for 5-char word with 115s left out of 120s, but got %d", expectedScore, lastScore)
	}

	// Test that score decreases as time decreases
	lastScore = 145
	for secondsLeft := 105; secondsLeft >= 5; secondsLeft -= 10 {
		newScore := calculateGuesserScore(5, secondsLeft, 120)
		if newScore > lastScore {
			t.Errorf("Score with more time taken should be lower. (LastScore: %d; NewScore: %d)", lastScore, newScore)
		}
		lastScore = newScore
	}
	
	// Test with 0 characters (edge case)
	zeroScore := calculateGuesserScore(0, 60, 120)
	expectedZeroScore := 0 + 100*60/120 // 0 + 50 = 50
	if zeroScore != expectedZeroScore {
		t.Errorf("Expected score %d for 0-char word with 60s left out of 120s, but got %d", expectedZeroScore, zeroScore)
	}
}

func Test_handleNameChangeEvent(t *testing.T) {
	lobby := &Lobby{}
	lobby.WriteJSON = func(player *Player, object interface{}) error {
		//Dummy to pass test.
		return nil
	}
	player := lobby.JoinPlayer("Kevin")

	handleNameChangeEvent(player, lobby, "Jim")

	expectedName := "Jim"
	if player.Name != expectedName {
		t.Errorf("playername didn't change; Expected %s, but was %s", expectedName, player.Name)
	}
}
