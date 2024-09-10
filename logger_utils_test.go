package main

import (
	"bufio"
	"bytes"
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
		conn    net.Conn
		ch      chan []byte
		shutdwn chan struct{}
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
	}{
		{
			name: "sending and receiving all the bytes of text1",
			args: args{
				conn:    nil,
				ch:      make(chan []byte, 10),
				shutdwn: make(chan struct{}, 1),
			},
			bytes:          text1,
			gotBytes:       []byte{},
			wantBytes:      text1,
			shutdwnDelay:   1 * time.Second,
			connCloseDelay: 1 * time.Second,
			shouldClose:    true,
			sigs:           make(chan os.Signal, 1),
			throwawayChan:  make(chan struct{}, 1),
		},
		{
			name: "sending and receiving all the bytes of text2",
			args: args{
				conn:    nil,
				ch:      make(chan []byte, 10),
				shutdwn: make(chan struct{}, 1),
			},
			bytes:          text2,
			gotBytes:       []byte{},
			wantBytes:      text2,
			shutdwnDelay:   1 * time.Second,
			connCloseDelay: 1 * time.Second,
			shouldClose:    true,
			sigs:           make(chan os.Signal, 1),
			throwawayChan:  make(chan struct{}, 1),
		},
		{
			name: "shutdwn signal should prevent any work being done",
			args: args{
				conn:    nil,
				ch:      make(chan []byte, 10),
				shutdwn: make(chan struct{}, 1),
			},
			bytes:          text2,
			gotBytes:       []byte{},
			wantBytes:      []byte{},
			shutdwnDelay:   0 * time.Second,
			connCloseDelay: 0 * time.Second,
			shouldClose:    true,
			sigs:           make(chan os.Signal, 1),
			throwawayChan:  make(chan struct{}, 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatal(err)
			}
			addr := l.Addr().String()

			go receiveAndAppend(t, &tt.gotBytes, tt.args.ch)
			go dialAndWrite(t, tt.bytes, addr, tt.connCloseDelay, tt.shouldClose)
			go shutdown(tt.throwawayChan, tt.args.shutdwn, tt.sigs)

			go func() {
				time.Sleep(tt.shutdwnDelay)
				tt.sigs <- syscall.SIGINT
			}()

			tt.args.conn, err = l.Accept()
			if err != nil {
				t.Fatal(err)
			}

			handleConn(tt.args.conn, tt.args.ch, tt.args.shutdwn)

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
			slog.Info("channel closed? breaking out of receiveAndAppend loop...")
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

func Test_log(t *testing.T) {
	type args struct {
		ch     chan []byte
		logger *lumberjack.Logger
	}
	tests := []struct {
		name  string
		args  args
		bytes []byte
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
			},
			bytes: []byte("this is simply one line 12345689</!)@(*&#%$()*&%#@0>\n"),
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
			},
			bytes: []byte("this\n is\njust\na\nbunch\nof\nlines\n12345689</!)@(*&#%$()*&%#@0>\n"),
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
			bytes: text1,
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
			bytes: text2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Remove(tt.args.logger.Filename)
			if err != nil {
				t.Fatal(err)
			}
			go log(tt.args.ch, tt.args.logger)

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

			time.Sleep(500 * time.Millisecond)

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
		stopLoggerLoop chan struct{}
		stopHandleConn chan struct{}
		sigs           chan os.Signal
	}
	tests := []struct {
		name   string
		args   args
		signal os.Signal
	}{
		{
			name: "sending SIGINT",
			args: args{
				stopLoggerLoop: make(chan struct{}, 1),
				stopHandleConn: make(chan struct{}, 1),
				sigs:           make(chan os.Signal, 1),
			},
			signal: syscall.SIGINT,
		},
		{
			name: "sending SIGTERM",
			args: args{
				stopLoggerLoop: make(chan struct{}, 1),
				stopHandleConn: make(chan struct{}, 1),
				sigs:           make(chan os.Signal, 1),
			},
			signal: syscall.SIGTERM,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := &sync.WaitGroup{}
			go func() {
				wg.Add(2)
				for {
					select {
					case <-tt.args.stopHandleConn:
						slog.Info("Test_shutdown: stopHandleConn channel got signal")
						wg.Done()
					case <-tt.args.stopLoggerLoop:
						slog.Info("Test_shutdown: stopLoggerLoop channel got signal")
						wg.Done()
					}
				}
			}()

			go func() {
				time.Sleep(500 * time.Millisecond)
				syscall.Kill(syscall.Getpid(), tt.signal.(syscall.Signal))
				// tt.args.sigs <- tt.signal
			}()

			shutdown(tt.args.stopLoggerLoop, tt.args.stopHandleConn, tt.args.sigs)
			// if we get here, shutdown has exited, which means it caught the int/term
			// signal
			wg.Wait()
			slog.Info("both channels received shutdown signal and shutdown func has returned")
		})
	}
}

func TestReadBytesWithTimeout(t *testing.T) {
	type args struct {
		r     BytesReader
		delim byte
		d     time.Duration
	}
	tests := []struct {
		name        string
		args        args
		want        []byte
		wantErr     error
		bytes       []byte
		readerDelay time.Duration
	}{
		{
			name:        "will timeout",
			args:        args{r: nil, delim: '\n', d: 250 * time.Millisecond},
			want:        nil,
			wantErr:     ReadTimeoutError,
			bytes:       text1,
			readerDelay: 500 * time.Millisecond,
		},
		{
			name:        "reading from text1 reader. shouldn't time out",
			args:        args{r: nil, delim: '\n', d: 500 * time.Millisecond},
			want:        []byte("Lorem ipsum dolor sit amet, consectetuer adipiscing elit.\n"),
			wantErr:     nil,
			bytes:       text1,
			readerDelay: 250 * time.Millisecond,
		},
		{
			name:        "reading from text2 reader. shouldn't time out",
			args:        args{r: nil, delim: '\n', d: 500 * time.Millisecond},
			want:        []byte("Lorem TEXT2 dolor sit amet, consectetuer adipiscing elit.\n"),
			wantErr:     nil,
			bytes:       text2,
			readerDelay: 250 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.bytes)
			bufioReader := bufio.NewReader(r)

			tt.args.r = &delayedReader{
				r: bufioReader,
				d: tt.readerDelay,
			}

			got, err := ReadBytesWithTimeout(tt.args.r, tt.args.delim, tt.args.d)

			if err != tt.wantErr {
				slog.Error("ERROR AREN'T EQUAL")
				t.Errorf("ReadBytesWithTimeout() got error %v. want error %v", err, tt.wantErr)
				return
			}

			if !bytes.Equal(got, tt.want) {
				slog.Error("BYTE SLICES AREN'T EQUAL")
				t.Errorf(
					"ReadBytesWithTimeout() got bytes %v. want bytes %v",
					string(got),
					string(tt.want),
				)
				return
			}
		})
	}
}

// here we really are only testing the timeout
func TestAcceptWithTimeout(t *testing.T) {
	type args struct {
		l net.Listener
		d time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    net.Conn // can't really test what i GET because tcpconn has unexported fields
		wantErr error
		delay   time.Duration
	}{
		{
			name: "conn is stablished before timeout",
			args: args{
				l: nil,
				d: 500 * time.Millisecond,
			},
			want:    nil,
			wantErr: nil,
			delay:   250 * time.Millisecond,
		},
		{
			name: "timeout is reached before conn",
			args: args{
				l: nil,
				d: 250 * time.Millisecond,
			},
			want:    nil,
			wantErr: NetTimeoutError,
			delay:   500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

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
				if tt.delay > tt.args.d { // if we're supposed to timeout
					return
				}
				time.Sleep(tt.delay)
				_, err := net.Dial("tcp", addr)
				if err != nil {
					t.Errorf("error dialing conn")
					return
				}
			}()
			got, err := AcceptWithTimeout(tt.args.l, tt.args.d)

			if err != tt.wantErr {
				t.Errorf("AcceptWithTimeout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				if _, ok := got.(*net.TCPConn); !ok {
					t.Errorf("AcceptWithTimeout() net.TCPConn assertion on got value failed")
					return
				}
			}
		})
	}
}
