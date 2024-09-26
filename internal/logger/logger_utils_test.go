package logger

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type delayedReader struct {
	r *bufio.Reader
	d time.Duration
}

// only for use in testing

func (r *delayedReader) ReadBytes(delim byte) ([]byte, error) {
	time.Sleep(r.d)
	b, err := r.r.ReadBytes(delim)
	return b, err
}

func Test_handleConn(t *testing.T) {
	type args struct {
		conn net.Conn
		ch   chan []byte
		ctx  context.Context
	}
	tests := []struct {
		name           string
		args           args
		bytes          []byte
		gotBytes       []byte
		wantBytes      []byte
		shutdwnDelay   time.Duration
		connCloseDelay time.Duration
		shouldClose    bool
		sigs           chan os.Signal
		throwawayChan  chan struct{}
		cancelFunc     context.CancelFunc
	}{
		{
			name: "sending and receiving all the bytes of text1",
			args: args{
				conn: nil,
				ch:   make(chan []byte, 10),
				ctx:  nil,
			},
			bytes:          text1,
			gotBytes:       []byte{},
			wantBytes:      text1,
			shutdwnDelay:   3 * time.Millisecond,
			connCloseDelay: 3 * time.Millisecond,
			shouldClose:    true,
			sigs:           make(chan os.Signal, 1),
			throwawayChan:  make(chan struct{}, 1),
			cancelFunc:     nil,
		},
		{
			name: "sending and receiving all the bytes of text2",
			args: args{
				conn: nil,
				ch:   make(chan []byte, 10),
				ctx:  nil,
			},
			bytes:          text2,
			gotBytes:       []byte{},
			wantBytes:      text2,
			shutdwnDelay:   3 * time.Millisecond,
			connCloseDelay: 3 * time.Millisecond,
			shouldClose:    true,
			sigs:           make(chan os.Signal, 1),
			throwawayChan:  make(chan struct{}, 1),
			cancelFunc:     nil,
		},
		{
			name: "shutdwn signal should prevent any work being done",
			args: args{
				conn: nil,
				ch:   make(chan []byte, 10),
				ctx:  nil,
			},
			bytes:          text2,
			gotBytes:       []byte{},
			wantBytes:      []byte{},
			shutdwnDelay:   0 * time.Second,
			connCloseDelay: 0 * time.Second,
			shouldClose:    true,
			sigs:           make(chan os.Signal, 1),
			throwawayChan:  make(chan struct{}, 1),
			cancelFunc:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatal(err)
			}
			addr := l.Addr().String()

			tt.args.ctx, tt.cancelFunc = context.WithCancel(context.Background())

			go receiveAndAppend(t, &tt.gotBytes, tt.args.ch)
			go dialAndWrite(t, tt.bytes, addr, tt.connCloseDelay, tt.shouldClose)
			go shutdown(tt.sigs, tt.cancelFunc)

			go func() {
				time.Sleep(tt.shutdwnDelay)
				tt.sigs <- syscall.SIGINT // TODO: run all tests with SIGTERM too
			}()

			tt.args.conn, err = l.Accept()
			if err != nil {
				t.Fatal(err)
			}

			handleConnWithCtx(tt.args.conn, tt.args.ch, tt.args.ctx)

			// post func checking
			if !bytes.Equal(tt.gotBytes, tt.wantBytes) {
				t.Errorf("got:\n%v\nwant:\n%v", string(tt.gotBytes), string(tt.wantBytes))
			}
		})
	}
}

func shutterDwn(t *testing.T, d time.Duration) {
	t.Helper()
	time.Sleep(d)
}

// (FOR TESTING ONLY) receives the messages from handleConn,
// and appends them to a slice, so we can check if anything was missed

func receiveAndAppend(t *testing.T, b *[]byte, ch <-chan []byte) {
	t.Helper()
	for {
		msg, ok := <-ch
		if !ok {
			slog.Info("receiveAndAppend(): channel closed? breaking out of loop...")
			break
		}
		*b = append(*b, msg...)
	}
}

// (FOR TESTING ONLY) connects via tcp and sends data.
// closes the connection when it's done.
func dialAndWrite(
	t *testing.T,
	b []byte,
	addr string,
	closingDelay time.Duration,
	shouldClose bool,
) {
	t.Helper()
	defer slog.Info("dialAndWrite has exited.")

	var (
		proto = "tcp"
		r     = bufio.NewReader(bytes.NewReader(b))
	)
	conn, err := net.Dial(proto, addr)
	if err != nil {
		t.Fatal(err)
		return
	}
	slog.Info("dialAndWrite(): stablished tcp conn")

	if shouldClose {
		defer func() {
			time.Sleep(closingDelay)
			conn.Close()
		}()
	}

	for {
		msg, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				slog.Info("reached EOF. checking byte slice...")
				if len(msg) == 0 {
					slog.Info("nothing left in byte slice after EOF. breaking")
					break
				}
				conn.Write(msg)
				break
			}
			t.Fatal(err)
		}
		conn.Write(msg)
	}
}

func Test_logWithCtx(t *testing.T) {
	type args struct {
		ch     chan []byte
		logger *lumberjack.Logger
		ctx    context.Context
	}
	tests := []struct {
		name   string
		args   args
		bytes  []byte
		cancel context.CancelFunc
	}{
		{
			name: "writing just one line",
			args: args{
				ch: make(chan []byte),
				logger: &lumberjack.Logger{
					Filename:   "testOneLine.log",
					MaxSize:    0,
					MaxAge:     0,
					MaxBackups: 0,
					LocalTime:  true,
					Compress:   false,
				},
				ctx: nil,
			},
			bytes:  []byte("this is simply one line 12345689</!)@(*&#%$()*&%#@0>\n"),
			cancel: nil,
		},
		{
			name: "writing a bunch of lines",
			args: args{
				ch: make(chan []byte),
				logger: &lumberjack.Logger{
					Filename:   "testBunchOfLines.log",
					MaxSize:    0,
					MaxAge:     0,
					MaxBackups: 0,
					LocalTime:  true,
					Compress:   false,
				},
				ctx: nil,
			},
			bytes:  []byte("this\n is\njust\na\nbunch\nof\nlines\n12345689</!)@(*&#%$()*&%#@0>\n"),
			cancel: nil,
		},
		{
			name: "writing 1 Lorem",
			args: args{
				ch: make(chan []byte),
				logger: &lumberjack.Logger{
					Filename:   "test1Lorem.log",
					MaxSize:    0,
					MaxAge:     0,
					MaxBackups: 0,
					LocalTime:  true,
					Compress:   false,
				},
			},
			bytes:  text1,
			cancel: nil,
		},
		{
			name: "writing 2 Lorem",
			args: args{
				ch: make(chan []byte),
				logger: &lumberjack.Logger{
					Filename:   "test2Lorem.log",
					MaxSize:    0,
					MaxAge:     0,
					MaxBackups: 0,
					LocalTime:  true,
					Compress:   false,
				},
			},
			bytes:  text2,
			cancel: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.ctx, tt.cancel = context.WithCancel(context.Background())
			defer os.Remove(tt.args.logger.Filename)
			go logWithCtx(tt.args.ch, tt.args.logger, tt.args.ctx)

			// f, err := os.Create(tt.args.logger.Filename)
			// if err != nil {
			// 	t.Fatal(err)
			// }

			bytesReader := bytes.NewReader(tt.bytes)

			r := bufio.NewReader(bytesReader)

			for {
				b, err := r.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						slog.Info("EOF reached while reading from bytes reader")
						tt.args.ch <- b
						break
					}
					slog.Error("non EOF error while reading from bytes reader", "error", err)
					t.Fatal(err)
				}
				tt.args.ch <- b
			}

			got, err := os.ReadFile(tt.args.logger.Filename)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(got, tt.bytes) {
				slog.Error("bytes not equal")
				t.Fatal(err)
			}
			slog.Info("bytes are equal")
			if string(got) != string(tt.bytes) {
				slog.Error("string not equal")
				t.Fatal(err)
			}
			slog.Info("strings are equal")
		})
	}
}

func Test_shutdown(t *testing.T) {
	type args struct {
		sigs   chan os.Signal
		cancel context.CancelFunc
	}
	tests := []struct {
		name   string
		args   args
		signal os.Signal
		ctx    context.Context
	}{
		{
			name: "sending SIGINT",
			args: args{
				sigs:   make(chan os.Signal, 1),
				cancel: nil,
			},
			signal: syscall.SIGINT,
			ctx:    nil,
		},
		{
			name: "sending SIGTERM",
			args: args{
				sigs:   make(chan os.Signal, 1),
				cancel: nil,
			},
			signal: syscall.SIGTERM,
			ctx:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ctx, tt.args.cancel = context.WithCancel(context.Background())
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-tt.ctx.Done()
				wg.Done()
			}()

			go shutdown(tt.args.sigs, tt.args.cancel)

			go func() {
				time.Sleep(5 * time.Millisecond) // to make sure our goroutine catches it
				syscall.Kill(syscall.Getpid(), tt.signal.(syscall.Signal))
			}()

			// if we get here, shutdown has exited, which means it caught the int/term
			// signal
			wg.Wait()
			slog.Info("both channels received shutdown signal and shutdown func has returned")
		})
	}
}

func TestReadBytesWithCtx(t *testing.T) {
	type args struct {
		r     BytesReader
		delim byte
		ctx   context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr error
		bytes   []byte
		cancel  context.CancelFunc
	}{
		{
			name:    "reading from text1 reader",
			args:    args{r: nil, delim: '\n', ctx: nil},
			want:    []byte("Lorem ipsum dolor sit amet, consectetuer adipiscing elit.\n"),
			wantErr: nil,
			bytes:   text1,
			cancel:  nil,
		},
		{
			name:    "reading from text2 reader",
			args:    args{r: nil, delim: '\n', ctx: nil},
			want:    []byte("Lorem TEXT2 dolor sit amet, consectetuer adipiscing elit.\n"),
			wantErr: nil,
			bytes:   text2,
			cancel:  nil,
		},
		{
			name:    "int before anything can get done",
			args:    args{r: nil, delim: '\n', ctx: nil},
			want:    []byte(""),
			wantErr: shutdownErr,
			bytes:   text2,
			cancel:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.r = bufio.NewReader(bytes.NewReader(tt.bytes))

			tt.args.ctx, tt.cancel = context.WithCancel(context.Background())

			if tt.wantErr == shutdownErr {
				tt.cancel()
			}

			got, err := ReadBytesWithCtx(tt.args.r, tt.args.delim, tt.args.ctx)

			if err != tt.wantErr {
				slog.Error("ERROR AREN'T EQUAL")
				t.Errorf("ReadBytesWithCtx() got error %v. want error %v", err, tt.wantErr)
				return
			}

			if !bytes.Equal(got, tt.want) {
				slog.Error("BYTE SLICES AREN'T EQUAL")
				t.Errorf(
					"ReadBytesWithCtx() got bytes %v. want bytes %v",
					string(got),
					string(tt.want),
				)
				return
			}
		})
	}
}

func TestAcceptWithCtx(t *testing.T) {
	type args struct {
		l   net.Listener
		ctx context.Context
	}
	tests := []struct {
		name           string
		args           args
		want           net.Conn // can't really test what i GET because tcpconn has unexported fields
		wantErr        error
		connDelay      time.Duration
		ctxCancelDelay time.Duration
		cancel         context.CancelFunc
	}{
		{
			name:           "conn is stablished before ctxCancel",
			args:           args{l: nil, ctx: nil},
			want:           nil,
			wantErr:        nil,
			connDelay:      1 * time.Millisecond,
			ctxCancelDelay: 3 * time.Millisecond,
			cancel:         nil,
		},
		{
			name:           "ctxCancel is reached before conn",
			args:           args{l: nil, ctx: nil},
			want:           nil,
			wantErr:        shutdownErr,
			connDelay:      3 * time.Millisecond,
			ctxCancelDelay: 1 * time.Millisecond,
			cancel:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			tt.args.ctx, tt.cancel = context.WithCancel(context.Background())

			defer func() { // cleaning up
				if tt.args.l != nil {
					tt.args.l.Close()
				}
			}()

			tt.args.l, err = net.Listen("tcp", ":0")
			if err != nil {
				t.Error(err)
				return
			}
			addr := tt.args.l.Addr().String()

			go func() {
				time.Sleep(tt.connDelay)
				_, err := net.Dial("tcp", addr)
				if err != nil {
					slog.Error("error dialing conn", "error", err)
					t.Errorf("error dialing conn")
					return
				}
			}()
			go func() {
				time.Sleep(tt.ctxCancelDelay)
				tt.cancel()
			}()
			got, err := AcceptWithCtx(tt.args.l, tt.args.ctx)

			if err != tt.wantErr {
				t.Errorf("AcceptWithCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				if _, ok := got.(*net.TCPConn); !ok {
					t.Errorf("AcceptWithCtx() net.TCPConn assertion on got value failed")
					return
				}
			}
		})
	}
}

var text2 []byte = []byte(`Lorem TEXT2 dolor sit amet, consectetuer adipiscing elit.
Vestibulum wisi massa, pulvinar vitae, vestibulum id, vestibulum et, erat.
Cras imperdiet.
Vivamus sed nunc sed pede tempor dictum.
Etiam at wisi sit amet nulla tincidunt mollis.
Sed eget TEXT2.
Cras sit amet massa id odio nonummy fringilla.
Aliquam eu velit et dolor varius egestas.
Phasellus congue.
Proin nec ante.
Phasellus vestibulum nulla semper lorem.
Sed tincidunt magna vitae nulla.
Sed sagittis congue risus.
Nam erat felis, rutrum non, ultricies et, nonummy vel, enim.
Praesent felis neque, venenatis et, hendrerit vitae, semper vel, erat.
Integer iaculis purus ut turpis.
Aliquam erat volutpat.
Praesent accumsan orci et odio.
Nullam metus dolor, venenatis a, sodales in, vestibulum in, enim.
Proin erat dui, pharetra ac, dapibus vitae, malesuada non, ante.
Integer sapien.
Praesent facilisis odio sit amet nunc.
In dapibus.
Integer tellus.
Phasellus ac tellus et quam ultricies volutpat.
Sed dolor orci, mattis ut, condimentum id, aliquet at, eros.
In sed nibh.
Maecenas eleifend commodo sem.
Nam eleifend eleifend leo.
Suspendisse magna ante, fringilla vel, euismod ac, rhoncus nec, enim.
Maecenas libero purus, tincidunt vel, dignissim ut, faucibus a, dui.
Proin non turpis vel dolor mattis facilisis.
Nullam condimentum, enim vitae volutpat varius, lacus dolor pellentesque urna, ac lobortis pede massa id felis.
Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus.
Sed convallis.
Aliquam vehicula urna eget enim.
Aenean odio massa, sollicitudin quis, rhoncus vel, cursus et, dolor.
Nulla pretium euismod lacus.
Vestibulum ante TEXT2 primis in faucibus orci luctus et ultrices posuere cubilia Curae; Cras est.
Donec tincidunt.
Fusce risus est, lacinia quis, interdum sed, adipiscing a, elit.
Fusce vel nunc eget est iaculis sagittis.
Morbi pulvinar.
Curabitur id ligula.
Sed nec orci at lorem lobortis laoreet.
Nam tincidunt euismod ligula.
Morbi ullamcorper tellus in lectus.
Sed elementum urna semper neque.
Donec luctus iaculis odio.
Nulla ultrices.
In tempor.
Morbi felis.
Vivamus sodales.
Phasellus sem.
Proin at massa quis arcu pretium mattis.
Cras vel velit.
Nam rutrum erat a risus.
Vestibulum lorem purus, imperdiet quis, vestibulum sit amet, posuere at, dolor.
Cras vehicula euismod ante.
Donec sagittis blandit purus.
Aliquam vitae leo eget orci dictum molestie.
Donec sit amet diam sed nunc pharetra elementum.
Quisque rutrum augue vel sapien.
Nam turpis.
Praesent diam leo, consequat vel, egestas nec, tempor ut, nisl.
Maecenas eu odio vel mi euismod tincidunt.
`)

var text1 []byte = []byte(`Lorem ipsum dolor sit amet, consectetuer adipiscing elit.
Vestibulum wisi massa, pulvinar vitae, vestibulum id, vestibulum et, erat.
Cras imperdiet.
Vivamus sed nunc sed pede tempor dictum.
Etiam at wisi sit amet nulla tincidunt mollis.
Sed eget ipsum.
Cras sit amet massa id odio nonummy fringilla.
Aliquam eu velit et dolor varius egestas.
Phasellus congue.
Proin nec ante.
Phasellus vestibulum nulla semper lorem.
Sed tincidunt magna vitae nulla.
Sed sagittis congue risus.
Nam erat felis, rutrum non, ultricies et, nonummy vel, enim.
Praesent felis neque, venenatis et, hendrerit vitae, semper vel, erat.
Integer iaculis purus ut turpis.
Aliquam erat volutpat.
Praesent accumsan orci et odio.
Nullam metus dolor, venenatis a, sodales in, vestibulum in, enim.
Proin erat dui, pharetra ac, dapibus vitae, malesuada non, ante.
Integer sapien.
Praesent facilisis odio sit amet nunc.
In dapibus.
Integer tellus.
Phasellus ac tellus et quam ultricies volutpat.
Sed dolor orci, mattis ut, condimentum id, aliquet at, eros.
In sed nibh.
Maecenas eleifend commodo sem.
Nam eleifend eleifend leo.
Suspendisse magna ante, fringilla vel, euismod ac, rhoncus nec, enim.
Maecenas libero purus, tincidunt vel, dignissim ut, faucibus a, dui.
Proin non turpis vel dolor mattis facilisis.
Nullam condimentum, enim vitae volutpat varius, lacus dolor pellentesque urna, ac lobortis pede massa id felis.
Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus.
Sed convallis.
Aliquam vehicula urna eget enim.
Aenean odio massa, sollicitudin quis, rhoncus vel, cursus et, dolor.
Nulla pretium euismod lacus.
Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Cras est.
Donec tincidunt.
Fusce risus est, lacinia quis, interdum sed, adipiscing a, elit.
Fusce vel nunc eget est iaculis sagittis.
Morbi pulvinar.
Curabitur id ligula.
Sed nec orci at lorem lobortis laoreet.
Nam tincidunt euismod ligula.
Morbi ullamcorper tellus in lectus.
Sed elementum urna semper neque.
Donec luctus iaculis odio.
Nulla ultrices.
In tempor.
Morbi felis.
Vivamus sodales.
Phasellus sem.
Proin at massa quis arcu pretium mattis.
Cras vel velit.
Nam rutrum erat a risus.
Vestibulum lorem purus, imperdiet quis, vestibulum sit amet, posuere at, dolor.
Cras vehicula euismod ante.
Donec sagittis blandit purus.
Aliquam vitae leo eget orci dictum molestie.
Donec sit amet diam sed nunc pharetra elementum.
Quisque rutrum augue vel sapien.
Nam turpis.
Praesent diam leo, consequat vel, egestas nec, tempor ut, nisl.
Maecenas eu odio vel mi euismod tincidunt.
`)
