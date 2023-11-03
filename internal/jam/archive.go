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

type CompressMode int

const (
	Uncompressed CompressMode = iota
	GZip
	BZip2
	XZ
	ZStd
)

const (
	DefaultCompressMode = GZip
)

type (
	TarOptions struct{}

	CompressOptions struct {
		Mode CompressMode
	}
)

func Tar(r io.Reader, dest string, opts *TarOptions) error {
	return nil
}

func UntarStream(r io.Reader, dest string, opts *TarOptions) error {
	return errors.New("not implemented")
}

func detectCompressMode(ctx context.Context, r io.Reader) (CompressMode, error) {
	return 0, errors.New("not implemented")
}

func DecompressStream(ctx context.Context, r io.Reader, dst io.Writer) error {
	mode, err := detectCompressMode(ctx, r)
	if err != nil {
		return err
	}

	opts := CompressOptions{
		Mode: mode,
	}

	return DecompressStreamWithOptions(ctx, r, dst, &opts)
}

func DecompressWithOptions(ctx context.Context, r io.Reader, dest string, opts *CompressOptions) error {
	f, err := os.Open(dest)
	if err != nil {
		return err
	}

	return DecompressStreamWithOptions(ctx, r, f, opts)
}

func DecompressStreamWithOptions(_ context.Context, src io.Reader, dest io.Writer, opts *CompressOptions) error {
	newReader, err := newCompressReader(context.Background(), src, opts)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dest, newReader); err != nil {
		return err
	}

	return nil
}

func newCompressReader(_ context.Context, src io.Reader, opts *CompressOptions) (io.ReadCloser, error) {
	if opts == nil {
		opts = &CompressOptions{
			Mode: Uncompressed,
		}
	}

	switch opts.Mode {
	case Uncompressed:
		return io.NopCloser(src), nil
	case GZip:
		return gzip.NewReader(src)
	case BZip2:
		return bzip2.NewReader(src, nil)
	case XZ:
		xzReader, err := xz.NewReader(src)
		if err != nil {
			return nil, err
		}

		return io.NopCloser(xzReader), nil
	case ZStd:
		zstdReader, err := zstd.NewReader(src)
		if err != nil {
			return nil, err
		}

		return io.NopCloser(zstdReader), nil
	default:
		return nil, errors.New("not supported")
	}
}

func CompressWithOptions(ctx context.Context, src io.Reader, dest string, opts *CompressOptions) error {
	f, err := os.Open(dest)
	if err != nil {
		return err
	}

	return CompressStreamWithOptions(ctx, src, f, opts)
}

func newCompressWriter(_ context.Context, src io.Writer, opts *CompressOptions) (io.Writer, error) {
	if opts == nil {
		opts = &CompressOptions{
			Mode: Uncompressed,
		}
	}

	switch opts.Mode {
	case Uncompressed:
		return src, nil
	case GZip:
		return gzip.NewWriter(src), nil
	case BZip2:
		return bzip2.NewWriter(src, nil)
	case XZ:
		return xz.NewWriter(src)
	case ZStd:
		return zstd.NewWriter(src)
	default:
		return nil, errors.New("not supported")
	}
}

func CompressStreamWithOptions(_ context.Context, src io.Reader, dst io.Writer, opts *CompressOptions) error {
	newWriter, err := newCompressWriter(context.Background(), dst, opts)
	if err != nil {
		return err
	}

	if _, err := io.Copy(newWriter, src); err != nil {
		return err
	}

	return nil
}
