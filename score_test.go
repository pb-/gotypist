package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const epsilon = 1e-6

func TestErrorScore(t *testing.T) {
	assert.InEpsilon(t, 100., errorScore("helloworld", 0), epsilon)
	assert.InEpsilon(t, 50., errorScore("helloworld", 1), epsilon)
	assert.InEpsilon(t, 100./3, errorScore("helloworld", 2), epsilon)

	assert.InEpsilon(t, 10., errorScore("ö", 0), epsilon)
}

func TestSpeedScore(t *testing.T) {
	assert.InEpsilon(t, 100., speedScore("helloworld", time.Duration(0)), epsilon)
	assert.Equal(t, 100./1.1, speedScore(
		"helloworld", time.Duration(time.Millisecond*100)), epsilon)
}

func TestScore(t *testing.T) {
	assert.InEpsilon(t, 0.2*100./1.1+0.8*50, score(
		"helloworld", time.Duration(time.Millisecond*100), 1), epsilon)
}
