// Copyright (c) Microsoft. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.
package main

// Adds automatic retry capability to HdfsAccessor with respect to RetryPolicy
type FaultTolerantHdfsAccessor struct {
	Impl        HdfsAccessor
	RetryPolicy *RetryPolicy
}

var _ HdfsAccessor = (*FaultTolerantHdfsAccessor)(nil) // ensure FaultTolerantHdfsAccessor implements HdfsAccessor

// Creates an instance of FaultTolerantHdfsAccessor
func NewFaultTolerantHdfsAccessor(impl HdfsAccessor, retryPolicy *RetryPolicy) *FaultTolerantHdfsAccessor {
	return &FaultTolerantHdfsAccessor{
		Impl:        impl,
		RetryPolicy: retryPolicy}
}

// Ensures HDFS accessor is connected to the HDFS name node
func (this *FaultTolerantHdfsAccessor) EnsureConnected() error {
	op := this.RetryPolicy.StartOperation()
	for {
		err := this.Impl.EnsureConnected()
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("Connect: %s", err) {
			return err
		}
	}
}

// Opens HDFS file for reading
func (this *FaultTolerantHdfsAccessor) OpenRead(path string) (HdfsReader, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		result, err := this.Impl.OpenRead(path)
		if err == nil {
			// wrapping returned HdfsReader with FaultTolerantHdfsReader
			return NewFaultTolerantHdfsReader(path, result, this.Impl, this.RetryPolicy), nil
		}
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] OpenRead: %s", path, err) {
			return nil, err
		}
	}
}

// Opens HDFS file for writing
func (this *FaultTolerantHdfsAccessor) OpenWrite(path string) (HdfsWriter, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		result, err := this.Impl.OpenWrite(path)
		if err == nil {
			// wrapping returned HdfsWriter with FaultTolerantHdfsReader
			return &FaultTolerantHdfsWriter{Impl: result}, nil
		}
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] OpenWrite: %s", path, err) {
			return nil, err
		}
	}
}

// Enumerates HDFS directory
func (this *FaultTolerantHdfsAccessor) ReadDir(path string) ([]Attrs, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		result, err := this.Impl.ReadDir(path)
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] ReadDir: %s", path, err) {
			return result, err
		}
	}
}

// Retrieves file/directory attributes
func (this *FaultTolerantHdfsAccessor) Stat(path string) (Attrs, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		result, err := this.Impl.Stat(path)
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] Stat: %s", path, err) {
			return result, err
		}
	}
}