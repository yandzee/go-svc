package identity

type Credentials map[string]string

type FieldCheck struct {
	IsCorrect bool   `json:"isCorrect"`
	Details   string `json:"details"`
}

type CredentialsCheck map[string]FieldCheck

func (fc CredentialsCheck) HasIncorrect() (*FieldCheck, bool) {
	for _, ch := range fc {
		if !ch.IsCorrect {
			return &ch, true
		}
	}

	return nil, false
}

func (f Credentials) Get(key string) string {
	return f[key]
}
