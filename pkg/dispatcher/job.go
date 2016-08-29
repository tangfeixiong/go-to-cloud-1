package dispatcher

// Job represents the job to be run
type Job struct {
	Payload Payload
}

// A buffered channel that we can send work requests on.
var JobQueue chan Job

type HandleFunc func()

type Payload struct {
	Handler HandleFunc
}

func (j Job) PayloadHandler() HandleFunc {
	return j.Payload.Handler
}
