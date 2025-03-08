package extract

type options struct {
	inputPath  string
	outputPath string
}

func (o *options) Validate() error {
	if o.inputPath == "" {
		return ErrInvalidInputPath
	}

	if o.outputPath == "" {
		return ErrInvalidOutputPath
	}

	return nil
}
