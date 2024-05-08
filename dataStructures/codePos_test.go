package dataStructures

import "testing"

func TestCombine(t *testing.T) {
	fileName := "main.neco"

	codePositions := []CodePos{
		{&fileName, 1, 1, 1, 10},
		{&fileName, 1, 1, 12, 25},

		{&fileName, 1, 1, 1, 1},
		{&fileName, 1, 1, 10, 10},

		{&fileName, 1, 1, 10, 10},
		{&fileName, 1, 1, 1, 1},

		{&fileName, 1, 1, 10, 10},
		{&fileName, 2, 2, 1, 1},

		{&fileName, 2, 2, 1, 1},
		{&fileName, 1, 1, 10, 10},

		{&fileName, 1, 1, 1, 10},
		{&fileName, 2, 3, 12, 25},

		{&fileName, 2, 3, 12, 25},
		{&fileName, 1, 1, 1, 10},
	}

	editedCodePostions := []CodePos{
		{&fileName, 1, 1, 1, 25},
		{&fileName, 1, 1, 1, 10},
		{&fileName, 1, 1, 1, 10},
		{&fileName, 1, 2, 10, 1},
		{&fileName, 1, 2, 10, 1},
		{&fileName, 1, 3, 1, 25},
		{&fileName, 1, 3, 1, 25},
	}

	for i := 0; i < len(codePositions); i += 2 {
		codePos := codePositions[i].Combine(&codePositions[i+1])

		if *codePos != editedCodePostions[i/2] {
			t.Errorf("(%s).Combine(%s): %s, want %s", codePositions[i], codePositions[i+1], codePos, editedCodePostions[i/2])
		}
	}
}

func TestCodePosString(t *testing.T) {
	fileName := "main.neco"

	codePositions := map[CodePos]string{
		{&fileName, 1, 1, 1, 25}:    "main.neco 1:1 1:25",
		{&fileName, 10, 14, 2, 1}:   "main.neco 10:2 14:1",
		{&fileName, 100, 100, 1, 1}: "main.neco 100:1 100:1",
		{&fileName, 5, 5, 5, 5}:     "main.neco 5:5 5:5",
		{&fileName, 1, 2, 1, 1}:     "main.neco 1:1 2:1",
		{&fileName, 1, 1, 1, 2}:     "main.neco 1:1 1:2",
		{&fileName, 1, 1, 2, 1}:     "main.neco 1:2 1:1",
		{&fileName, 2, 1, 1, 1}:     "main.neco 2:1 1:1",
	}

	for codePos, str := range codePositions {
		if codePos.String() != str {
			t.Errorf("(%s).String(): %s, want %s", codePos, codePos.String(), str)
		}
	}
}
