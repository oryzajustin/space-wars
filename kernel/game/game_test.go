package game

import (
	"testing"

	"github.com/davidhorak/space-wars/kernel/physics"
	"github.com/davidhorak/space-wars/kernel/physics/collider"
	"github.com/stretchr/testify/assert"
)

type MockGameObject struct {
	position physics.Vector2
}

func (object *MockGameObject) ID() int64 {
	return 0
}

func (object *MockGameObject) Enabled() bool {
	return true
}

func (object *MockGameObject) SetEnabled(enabled bool) {
}

func (object *MockGameObject) Position() physics.Vector2 {
	return object.position
}

func (object *MockGameObject) SetPosition(position physics.Vector2) {
	object.position = position
}

func (object *MockGameObject) Update(deltaTimeMs float64, gameManager *GameManager) {
	object.position = object.position.Add(physics.Vector2{X: 1 / deltaTimeMs, Y: 1 / deltaTimeMs})
}

func (object *MockGameObject) Collider() collider.Collider {
	return nil
}

func (object *MockGameObject) OnCollision(other GameObject, gameManager *GameManager, order int) {
}

func (object *MockGameObject) Serialize() map[string]interface{} {
	return nil
}

func TestNewGame(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)

	assert.Equal(t, int64(1234567890), game.seed)
	assert.Equal(t, Initialized, game.status)
	assert.Equal(t, physics.Size{Width: 1024, Height: 768}, game.size)
	assert.GreaterOrEqual(t, len(game.manager.GameObjects()), MinAsteroids)
}

func TestGame_Status(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)
	assert.Equal(t, Initialized, game.Status())

	game.Start()
	assert.Equal(t, Running, game.Status())
}

func TestGame_Start(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)
	assert.Equal(t, Initialized, game.Status())

	game.Start()
	assert.Equal(t, Running, game.Status())
	assert.Equal(t, "Game state changed to: running", game.manager.Logger().Logs()[0].message)
}

func TestGame_Pause(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)
	assert.Equal(t, Initialized, game.Status())

	game.Start()
	assert.Equal(t, Running, game.Status())

	game.Pause()
	assert.Equal(t, Paused, game.Status())
	assert.Equal(t, "Game state changed to: paused", game.manager.Logger().Logs()[1].message)
}

func TestGame_Reset(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)
	assert.Equal(t, Initialized, game.Status())

	game.Start()
	assert.Equal(t, Running, game.Status())

	game.Reset()
	assert.Equal(t, Running, game.Status())
	assert.Equal(t, 0, len(game.manager.Logger().Logs()))
}

func TestGame_Update(t *testing.T) {
	t.Run("Updates game object positions", func(t *testing.T) {
		game := NewGame(physics.Size{Width: 1000, Height: 1000}, 1234567890)
		gameObject := &MockGameObject{
			position: physics.Vector2{X: 100, Y: 100},
		}
		game.manager.AddGameObject(gameObject)

		game.Update(100)

		assert.InDelta(t, 100.01, gameObject.Position().X, 0.01)
		assert.InDelta(t, 100.01, gameObject.Position().Y, 0.01)
	})

	t.Run("Wraps objects around screen edges", func(t *testing.T) {
		game := NewGame(physics.Size{Width: 1000, Height: 1000}, 1234567890)
		gameObject := &MockGameObject{
			position: physics.Vector2{X: 1000, Y: 1000},
		}
		game.manager.AddGameObject(gameObject)

		game.Update(100)
		game.Update(100)

		assert.InDelta(t, 0, gameObject.Position().X, 0.1)
		assert.InDelta(t, 0, gameObject.Position().Y, 0.1)

		game.Update(-25)
		game.Update(-25)

		assert.InDelta(t, 1000, gameObject.Position().X, 0.1)
		assert.InDelta(t, 1000, gameObject.Position().Y, 0.1)
	})

	t.Run("Handles collisions between objects", func(t *testing.T) {
		game := NewGame(physics.Size{Width: 1000, Height: 1000}, 1234567890)
		spaceship := NewSpaceship(NewUUID(), "test", physics.Vector2{X: 100, Y: 100}, 0)
		asteroid := NewAsteroid(NewUUID(), physics.Vector2{X: 150, Y: 100}, 50)
		game.manager.AddGameObjects([]GameObject{spaceship, asteroid})

		game.Update(100)

		assert.False(t, spaceship.Enabled())
		assert.Equal(t, float64(0), spaceship.health)
		assert.True(t, asteroid.Enabled())
	})

	t.Run("Ignores disabled objects", func(t *testing.T) {
		game := NewGame(physics.Size{Width: 1000, Height: 1000}, 1234567890)
		spaceship := NewSpaceship(NewUUID(), "test", physics.Vector2{X: 100, Y: 100}, 0)
		asteroid := NewAsteroid(NewUUID(), physics.Vector2{X: 150, Y: 100}, 50)
		game.manager.AddGameObjects([]GameObject{spaceship, asteroid})

		spaceship.SetEnabled(false)

		game.Update(100)

		assert.Equal(t, float64(100), spaceship.health)
		assert.True(t, asteroid.Enabled())
	})

	t.Run("Game ends when manager ends", func(t *testing.T) {
		game := NewGame(physics.Size{Width: 1000, Height: 1000}, 1234567890)
		game.Start()

		game.Update(100) // 100ms

		assert.Equal(t, Ended, game.Status())
		assert.Equal(t, "Game state changed to: ended", game.manager.Logger().Logs()[1].message)
	})
}

func TestGame_SpaceshipAction(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)
	game.AddSpaceship("test", physics.Vector2{X: 100, Y: 100}, 0)

	game.SpaceshipAction("test", func(spaceShip *Spaceship, gameManager *GameManager) {
		spaceShip.position = physics.Vector2{X: 200, Y: 200}
	})

	spaceship, err := game.manager.GetSpaceship("test")
	assert.NoError(t, err)
	assert.Equal(t, physics.Vector2{X: 200, Y: 200}, spaceship.position)
}

func TestGame_AddSpaceship(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)
	game.AddSpaceship("test", physics.Vector2{X: 100, Y: 100}, 0)

	gameObjects := game.manager.GameObjects()
	spaceship := gameObjects[len(gameObjects)-1].(*Spaceship)
	assert.IsType(t, &Spaceship{}, spaceship)
	assert.Equal(t, "test", spaceship.name)
}

func TestGame_RemoveSpaceship(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)
	game.AddSpaceship("test", physics.Vector2{X: 100, Y: 100}, 0)

	game.RemoveSpaceship("test")

	gameObjects := game.manager.GameObjects()
	assert.IsType(t, &Asteroid{}, gameObjects[len(gameObjects)-1])
}

func TestGame_Serialize(t *testing.T) {
	game := NewGame(physics.Size{Width: 1024, Height: 768}, 1234567890)
	game.AddSpaceship("test", physics.Vector2{X: 100, Y: 100}, 0)

	serialized := game.Serialize()
	assert.Equal(t, "initialized", serialized["status"])
	assert.Equal(t, int64(1234567890), serialized["seed"])
	assert.Equal(t, 1024.0, serialized["size"].(map[string]interface{})["width"])
	assert.Equal(t, 768.0, serialized["size"].(map[string]interface{})["height"])
	assert.GreaterOrEqual(t, len(serialized["gameObjects"].([]interface{})), MinAsteroids)
	assert.Equal(t, 0, len(serialized["logs"].([]interface{})))
}