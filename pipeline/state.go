package pipeline

import "github.com/yandzee/go-svc/flow"

type Stage[A any] interface {
	Id() string

	Act(A) Result[A]

	OnEnter()
	OnExit(Result[A])
}

type Result[A any] struct {
	Stage   Stage[A]     `json:"state"`
	Control flow.Control `json:"control"`
	Args    A            `json:"args"`
}

func Run[A any](init Stage[A], args A) Result[A] {
	result := Result[A]{
		Stage:   init,
		Control: flow.Continue,
		Args:    args,
	}

	if init == nil {
		result.Control = flow.Break
		return result
	}

	lastStateId := init.Id()
	init.OnEnter()

	for result.Control == flow.Continue {
		stage := result.Stage
		id := stage.Id()

		if lastStateId != id {
			stage.OnEnter()
			lastStateId = id
		}

		result = stage.Act(result.Args)

		if result.Stage == nil {
			result.Control = flow.Break
			break
		}

		if result.Stage.Id() != id {
			stage.OnExit(result)
		}
	}

	return result
}

func Continue[A any](s Stage[A], args A) Result[A] {
	return Result[A]{
		Stage:   s,
		Control: flow.Continue,
		Args:    args,
	}
}

func Break[A any](s Stage[A], args A) Result[A] {
	return Result[A]{
		Stage:   s,
		Control: flow.Break,
		Args:    args,
	}
}
