package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const epsilon = 1e-6

func TestErrorScore(t *testing.T) {
	assert.InEpsilon(t, 1., errorScore("helloworld", 0), epsilon)
	assert.InEpsilon(t, 1./2, errorScore("helloworld", 1), epsilon)
	assert.InEpsilon(t, 1./3, errorScore("helloworld", 2), epsilon)
}

func TestSpeedScore(t *testing.T) {
	assert.InEpsilon(t, 1., speedScore("helloworld", time.Duration(0)), epsilon)
	assert.InEpsilon(t, 5./6, speedScore("helloworld", time.Duration(1*time.Second)), epsilon)
	assert.InEpsilon(t, 5./7, speedScore("helloworld", time.Duration(2*time.Second)), epsilon)
	assert.InEpsilon(t, 50./51, speedScore(
		"helloworld", time.Duration(time.Millisecond*100)), epsilon)

	assert.InEpsilon(t, 1./2, speedScore("ö", time.Duration(500*time.Millisecond)), epsilon)
}

func TestScore(t *testing.T) {
	assert.InEpsilon(t, 0.2*50./51+0.8*0.5, score(
		"helloworld", time.Duration(time.Millisecond*100), 1), epsilon)
}

func TestLevel(t *testing.T) {
	assert.Equal(t, 42, level(requiredScore(42)))
}
