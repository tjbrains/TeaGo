package caches_test

import (
	"github.com/tjbrains/TeaGo/caches"
	"testing"
	"time"
)

func TestNewFactory(t *testing.T) {
	var factory = caches.NewFactory()
	factory.Set("hello", "world").ExpireAt(time.Now().Add(10 * time.Second))

	value, found := factory.Get("hello")
	if !found {
		t.Fatal("[ERROR]", "'hello' not found")
	}

	if value != "world" {
		t.Fatal("[ERROR]", "'hello' not equal 'world'")
	}

	t.Log("ok")
}

func TestNewFactory_Clean(t *testing.T) {
	var factory = caches.NewFactory()
	factory.Set("hello", "world").ExpireAt(time.Now().Add(-10 * time.Second))
	t.Log(len(factory.Items()))
	factory.Clean()
	t.Log(len(factory.Items()))
}

func TestNewFactory_CleanLoop(t *testing.T) {
	var factory = caches.NewFactoryInterval(1 * time.Second)
	factory.Set("hello", "world").ExpireAt(time.Now().Add(2 * time.Second))

	time.Sleep(3 * time.Second)

	t.Log(time.Now())
	t.Log(len(factory.Items()))
}
