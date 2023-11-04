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

type ArchiveMode int

const (
	NopArchive ArchiveMode = iota
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

	ArchiveOptions struct {
		Mode ArchiveMode
	}

	UnarchiveFunc func(context.Context, io.Reader) (io.ReadCloser, error)

	CompressFunc func(context.Context, io.Writer) (io.Writer, error)
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

func findArchiveMode(ctx context.Context, r io.Reader) (ArchiveMode, error) {
	return 0, errors.New("not implemented")
}

func unarchiveStream(ctx context.Context, r io.Reader, dst io.Writer) error {
	mode, err := findArchiveMode(ctx, r)
	if err != nil {
		return err
	}

	opts := ArchiveOptions{
		Mode: mode,
	}

	return UnarchiveStreamWithOptions(ctx, r, dst, &opts)
}

func UnarchiveWithOptions(ctx context.Context, r io.Reader, dest string, opts *ArchiveOptions) error {
	f, err := os.Open(dest)
	if err != nil {
		return err
	}

	return UnarchiveStreamWithOptions(ctx, r, f, opts)
}

func UnarchiveStreamWithOptions(ctx context.Context, src io.Reader, dest io.Writer, opts *ArchiveOptions) error {
	decompressFunc := buildUnarchiveFunc(opts)

	r, err := decompressFunc(ctx, src)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dest, r); err != nil {
		return err
	}

	return nil
}

func buildUnarchiveFunc(opts *ArchiveOptions) UnarchiveFunc {
	if opts == nil {
		opts = &ArchiveOptions{
			Mode: NopArchive,
		}
	}

	switch opts.Mode {
	case Gzip:
		return gzipUnarchiver
	case Bzip2:
		return bzip2Unarchiver
	case Xz:
		return xzUnarchiver
	case Zstd:
		return zstdUnarchiver
	default:
		return nopUnarchiver
	}
}

func nopUnarchiver(_ context.Context, src io.Reader) (io.ReadCloser, error) {
	return io.NopCloser(src), nil
}

func gzipUnarchiver(_ context.Context, src io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(src)
}

func bzip2Unarchiver(_ context.Context, src io.Reader) (io.ReadCloser, error) {
	return bzip2.NewReader(src, nil)
}

func xzUnarchiver(_ context.Context, src io.Reader) (io.ReadCloser, error) {
	r, err := xz.NewReader(src)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(r), nil
}

func zstdUnarchiver(_ context.Context, src io.Reader) (io.ReadCloser, error) {
	r, err := zstd.NewReader(src)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(r), nil
}

func ArchiveWithOptions(ctx context.Context, src io.Reader, dest string, opts *ArchiveOptions) error {
	f, err := os.Open(dest)
	if err != nil {
		return err
	}

	return ArchiveStreamWithOptions(ctx, src, f, opts)
}

func ArchiveStreamWithOptions(ctx context.Context, src io.Reader, dst io.Writer, opts *ArchiveOptions) error {
	compress := buildArchiveFunc(opts)

	w, err := compress(ctx, dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, src); err != nil {
		return err
	}

	return nil
}

func buildArchiveFunc(opts *ArchiveOptions) CompressFunc {
	if opts == nil {
		opts = &ArchiveOptions{
			Mode: NopArchive,
		}
	}

	switch opts.Mode {
	case Gzip:
		return gzipArchiver
	case Bzip2:
		return bzip2Archiver
	case Xz:
		return xzArchiver
	case Zstd:
		return zstdArchiver
	default:
		return nopArchiver
	}
}

func nopArchiver(_ context.Context, src io.Writer) (io.Writer, error) {
	return src, nil
}

func gzipArchiver(_ context.Context, src io.Writer) (io.Writer, error) {
	return gzip.NewWriter(src), nil
}

func bzip2Archiver(_ context.Context, src io.Writer) (io.Writer, error) {
	return bzip2.NewWriter(src, nil)
}

func xzArchiver(_ context.Context, src io.Writer) (io.Writer, error) {
	return xz.NewWriter(src)
}

func zstdArchiver(_ context.Context, src io.Writer) (io.Writer, error) {
	return zstd.NewWriter(src)
}
