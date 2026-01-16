package utils

func RemoveJob(jobs []Job, job Job) []Job {
	for i, j := range jobs {
		if j.ID == job.ID {
			return append(jobs[:i], jobs[i+1:]...)
		}
	}
	return jobs
}
