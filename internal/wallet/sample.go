package wallet

import "github.com/rinor/jorcli/jnode"

func SampleWallets() []Wallet {
	return []Wallet{
		{
			Note:      "DAEDALUS",
			Mnemonics: "tired owner misery large dream glad upset welcome shuffle eagle pulp time",
			Funds: []jnode.InitialFund{
				{
					Address: "DdzFFzCqrhsktawSMCWJJy3Dpp9BCjYPVecgsMb5U2G7d1ErUUmwSZvfSY3Yjn5njNadfwvebpVNS5cD4acEKSQih2sR76wx2kF4oLXT",
					Value:   10_000,
				},
				{
					Address: "DdzFFzCqrhsg7eQHEfFE7cH7bKzyyUEKSoSiTmQtxAGnAeCW3pC2LXyxaT8T5sWH4zUjfjffik6p9VdXvRfwJgipU3tgzXhKkMDLt1hR",
					Value:   10_100,
				},
				{
					Address: "DdzFFzCqrhsw7G6njwb8FTBxVCh9GtB7RFvvz7KPNkHxeHtDwAPT2Y6QLDLxVCu7NNUQmwpAfgG5ZeGQkoWjrkbHPUeU9wzG3YFpohse",
					Value:   1,
				},
			},
		},
		{
			Note:      "DAEDALUS",
			Mnemonics: "edge club wrap where juice nephew whip entry cover bullet cause jeans",
			Funds: []jnode.InitialFund{
				{
					Address: "DdzFFzCqrhsf2sWcZLzXhyLoLZcmw3Zf3UcJ2ozG1EKTwQ6wBY1wMG1tkXtPvEgvE5PKUFmoyzkP8BL4BwLmXuehjRHJtnPj73E5RPMx",
					Value:   20_000,
				},
				{
					Address: "DdzFFzCqrhsogWSfcp4Dq9W1bcMzt86276PbDfzAKZxDhi3g6w6fRu6zYMT36uG8p3j8bCgsx4frkB3QH8m8ubUhAKRG5c8SLnGVTBh9",
					Value:   20_100,
				},
				{
					Address: "DdzFFzCqrhtDFbFvtrm3hhHuWUPY9ozkCW5JzuL4TcrXKMruWCrCSRzpc4mkWBUugPAGLesJv3ert9BH1cQJqXq2f4UN83WP5AZZN4jQ",
					Value:   20_200,
				},
				{
					Address: "sxtitePxjp5r4GxbM6EtS1EEe45zGoR4XDYnXYb9MuoE1HnoqDtKpRpdx4WjayaR72p2MKHFExAyDL89mJMoJ22WQR",
					Value:   20_300,
				},
				{
					Address: "sxtitePxjp5WJkHH5L6YWA5ZTRc8yEpLd9NYu3rMAFrVzfzWjAtkRPZ8UZHYzDjsigGijsFJ2iB6PFDvWdRYfCra66",
					Value:   20_400,
				},
				{
					Address: "sxtitePxjp5txDrVJU8cqwjDkAqx5odRt7kpMzVyXUQmEZL7wCA5fs29MJLCdux1Uz41xX1KTG5vqCHHXegidwnfFL",
					Value:   1,
				},
			},
		},
		{
			Note:      "YOROI",
			Mnemonics: "neck bulb teach illegal soul cry monitor claw amount boring provide village rival draft stone",
			Funds: []jnode.InitialFund{
				{
					Address: "Ae2tdPwUPEZ8og5u4WF5rmSyme5Gvp8RYiLM2u7Vm8CyDQzLN3VYTN895Wk",
					Value:   30_000,
				},
				{
					Address: "Ae2tdPwUPEZEAjEsQsCtBMkLKANxQUEvzLkumPWWYugLeXcgkeMCDH1gnuL",
					Value:   1,
				},
			},
		},
		{
			Note:      "DAEDALUS-PAPER",
			Mnemonics: "town lift more follow chronic lunch weird uniform earth census proof cave gap fancy topic year leader phrase state circle cloth reward dish survey act punch bounce",
			Funds: []jnode.InitialFund{
				{
					Address: "DdzFFzCqrhtCvPjBLTJKJdNWzfhnJx3967QEcuhhm1PQ2ca13fNNMh5KZentH5aWLysjEBc1rKDYMS3noNKNyxdCL8NHUZznZj9gofQJ",
					Value:   40_000,
				},
			},
		},
	}
}

func SampleWalletPaper() []Wallet {
	return []Wallet{
		{
			Note:      "DAEDALUS-PAPER",
			Mnemonics: "town lift more follow chronic lunch weird uniform earth census proof cave gap fancy topic year leader phrase state circle cloth reward dish survey act punch bounce",
			Funds: []jnode.InitialFund{
				{
					Address: "DdzFFzCqrhtCvPjBLTJKJdNWzfhnJx3967QEcuhhm1PQ2ca13fNNMh5KZentH5aWLysjEBc1rKDYMS3noNKNyxdCL8NHUZznZj9gofQJ",
					Value:   40_000,
				},
			},
		},
	}
}
