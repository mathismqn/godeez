package downloader

func handleError(err error) downloadResult {
	if path, ok := IsSkipError(err); ok {
		return downloadResult{
			skipped: true,
			path:    path,
		}
	}

	return downloadResult{
		success: false,
		err:     err,
	}
}
