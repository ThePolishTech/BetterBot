package messages

func FillResponses() {
	responses = []messageResponse{
		{triggers: []string{"promotion"}, response: "So when do I get a promotion?", isMedia: false},
		{triggers: []string{"warcrime", "war crime", "war-crime"}, response: "\"Geneva Convention\" has been added on the To-do-list", isMedia: false},
		{triggers: []string{"horny", "porn", "lewd"}, response: "# I shall grill all horny people\nhttps://tenor.com/bFz07.gif", isMedia: false},
		{triggers: []string{"choccy milk"}, response: "Pilot, I have acquired the choccy milk!", isMedia: false},
		{triggers: []string{"dead", "died"}, response: "F", isMedia: false},
		{triggers: []string{"doot"}, response: "https://tenor.com/tyG1.gif"},
		{triggers: []string{"sus", "among us", "amogus", "impostor"}, response: "Funny Amogus sussy impostor\nhttps://tenor.com/bs8aU.gif", isMedia: false},
		{triggers: []string{"mlik"}, response: "https://tenor.com/q6vqHU4ETLK.gif", isMedia: false},
		{triggers: []string{"risk-of-rain"}, response: "https://tenor.com/view/risk-of-rain-risk-of-rain-2-tater-brigade-rain-flying-gif-8263622248996575525", isMedia: false },

		{triggers: []string{"scronch", "scornch"}, response: "scronch", isMedia: true},
	}
}
