package jam

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"

	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

type CompressionMode int

const (
	Uncompressed CompressionMode = iota
	Gzip
	Bzip2
	Xz
	Zstd
)

const (
	DefaultCompressMode = Gzip
)

type (
	TarOptions struct {
		UseCLI bool
	}

	CompressionOptions struct {
		Mode CompressionMode
	}
)

func Tar(_ io.Reader, _ string, opts *TarOptions) error {
	if opts == nil {
		opts = &TarOptions{
			UseCLI: false,
		}
	}

	if opts.UseCLI {
	}

	return nil
}

func UntarStream(r io.Reader, dest string, opts *TarOptions) error {
	return errors.New("not implemented")
}

func findCompressionMode(ctx context.Context, r io.Reader) (CompressionMode, error) {
	return 0, errors.New("not implemented")
}

func DecompressStream(ctx context.Context, r io.Reader, dst io.Writer) error {
	mode, err := findCompressionMode(ctx, r)
	if err != nil {
		return err
	}

	opts := CompressionOptions{
		Mode: mode,
	}

	return DecompressStreamWithOptions(ctx, r, dst, &opts)
}

func DecompressWithOptions(ctx context.Context, r io.Reader, dest string, opts *CompressionOptions) error {
	f, err := os.Open(dest)
	if err != nil {
		return err
	}

	return DecompressStreamWithOptions(ctx, r, f, opts)
}

func DecompressStreamWithOptions(_ context.Context, src io.Reader, dest io.Writer, opts *CompressionOptions) error {
	newReader, err := newCompressReader(context.Background(), src, opts)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dest, newReader); err != nil {
		return err
	}

	return nil
}

func newCompressReader(_ context.Context, src io.Reader, opts *CompressionOptions) (io.ReadCloser, error) {
	if opts == nil {
		opts = &CompressionOptions{
			Mode: Uncompressed,
		}
	}

	switch opts.Mode {
	case Uncompressed:
		return io.NopCloser(src), nil
	case Gzip:
		return gzip.NewReader(src)
	case Bzip2:
		return bzip2.NewReader(src, nil)
	case Xz:
		xzReader, err := xz.NewReader(src)
		if err != nil {
			return nil, err
		}

		return io.NopCloser(xzReader), nil
	case Zstd:
		zstdReader, err := zstd.NewReader(src)
		if err != nil {
			return nil, err
		}

		return io.NopCloser(zstdReader), nil
	default:
		return nil, errors.New("not supported")
	}
}

func CompressWithOptions(ctx context.Context, src io.Reader, dest string, opts *CompressionOptions) error {
	f, err := os.Open(dest)
	if err != nil {
		return err
	}

	return CompressStreamWithOptions(ctx, src, f, opts)
}

func newCompressWriter(_ context.Context, src io.Writer, opts *CompressionOptions) (io.Writer, error) {
	if opts == nil {
		opts = &CompressionOptions{
			Mode: Uncompressed,
		}
	}

	switch opts.Mode {
	case Uncompressed:
		return src, nil
	case Gzip:
		return gzip.NewWriter(src), nil
	case Bzip2:
		return bzip2.NewWriter(src, nil)
	case Xz:
		return xz.NewWriter(src)
	case Zstd:
		return zstd.NewWriter(src)
	default:
		return nil, errors.New("not supported")
	}
}

func CompressStreamWithOptions(_ context.Context, src io.Reader, dst io.Writer, opts *CompressionOptions) error {
	newWriter, err := newCompressWriter(context.Background(), dst, opts)
	if err != nil {
		return err
	}

	if _, err := io.Copy(newWriter, src); err != nil {
		return err
	}

	return nil
}
