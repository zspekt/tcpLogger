package logger

import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/zspekt/tcpLogger/internal/setup"
)

func Test_logger(t *testing.T) {
	tests := []struct {
		name          string
		arg           *setup.Cfg
		bytes         []byte
		wantBytes     []byte
		shutdownDelay time.Duration
		signal        syscall.Signal
		//
		lineStop    int
		lineStopSig chan struct{}
	}{
		{
			name: "succesfully logging text3. ending with SIGINT",
			arg: &setup.Cfg{
				Port: "",
				Logger: &lumberjack.Logger{
					Filename:   "text3_test.txt",
					MaxSize:    0,
					MaxAge:     0,
					MaxBackups: 0,
					LocalTime:  true,
					Compress:   false,
				},
				Protocol: "tcp",
				Address:  "",
			},
			bytes:         text3,
			wantBytes:     text3,
			shutdownDelay: 1000 * time.Millisecond,
			signal:        syscall.SIGINT,
			lineStop:      5,
			lineStopSig:   make(chan struct{}),
		},
		{
			name: "succesfully logging text4. ending with SIGINT",
			arg: &setup.Cfg{
				Port: "",
				Logger: &lumberjack.Logger{
					Filename:   "text4_test.txt",
					MaxSize:    0,
					MaxAge:     0,
					MaxBackups: 0,
					LocalTime:  true,
					Compress:   false,
				},
				Protocol: "tcp",
				Address:  "",
			},
			bytes:         text4,
			wantBytes:     text4,
			shutdownDelay: 1000 * time.Millisecond,
			signal:        syscall.SIGINT,
			lineStop:      7,
			lineStopSig:   make(chan struct{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := net.Listen("tcp", "localhost:0") // just doing this to get an available port
			if err != nil {
				t.Fatal(err)
			}
			addr := strings.Split(l.Addr().String(), ":")
			l.Close()
			tt.arg.Address = addr[0]
			tt.arg.Port = addr[1]

			// we start a func that will send an INT/TERM sig, which shutdown()
			// (called concurrently by Run() ) will catch and then notify
			// both handleConn() and Run(), leading to a graceful shutdown.
			go func() {
				time.Sleep(tt.shutdownDelay)
				syscall.Kill(syscall.Getpid(), tt.signal)
			}()

			go dialAndWrite(t, tt.bytes, addr[0]+":"+addr[1], 0, true)

			Run(tt.arg)

			got, err := os.ReadFile(tt.arg.Logger.Filename)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				os.Remove(tt.arg.Logger.Filename)
			})

			if !bytes.Equal(got, tt.wantBytes) {
				t.Errorf("got:\n<<%v>>\nwant:\n<<%v>>", string(got), string(tt.wantBytes))
			}
		})

		t.Run("stopping mid exec:"+tt.name, func(t *testing.T) {
			l, err := net.Listen("tcp", "localhost:0") // just doing this to get an available port
			if err != nil {
				t.Fatal(err)
			}

			addr := strings.Split(l.Addr().String(), ":")
			l.Close()
			tt.arg.Address = addr[0]
			tt.arg.Port = addr[1]

			go func() {
				<-tt.lineStopSig
				time.Sleep(50 * time.Millisecond)
				syscall.Kill(syscall.Getpid(), tt.signal)
			}()

			go func(t *testing.T, b []byte, addr string) {
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

				for i := 0; i < tt.lineStop; i++ {
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
				tt.lineStopSig <- struct{}{}
			}(
				t,
				tt.bytes,
				addr[0]+":"+addr[1],
			)

			Run(tt.arg)

			got, err := os.ReadFile(tt.arg.Logger.Filename)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				os.Remove(tt.arg.Logger.Filename)
			})

			split := bytes.Split(tt.bytes, []byte("\n"))[:tt.lineStop]

			// we add a newline at the end since that's the format we want
			want := append(bytes.Join(split, []byte("\n")), '\n')

			if !bytes.Equal(got, want) {
				t.Errorf("got:\n<<%v>>\nwant:\n<<%v>>", string(got), string(want))
			}
		})
	}
}

var text4 []byte = []byte(`Lorem TEXT2 dolor sit amet, consectetuer adipiscing elit.
Vestibulum wisi massa, pulvinar vitae, vestibulum id, vestibulum et, erat.
Cras imperdiet.
Vivamus sed nunc sed pede tempor dictum.
Etiam at wisi sit amet nulla tincidunt mollis.
Sed eget TEXT4.
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
Vestibulum ante TEXT4 primis in faucibus orci luctus et ultrices posuere cubilia Curae; Cras est.
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

var text3 []byte = []byte(`Lorem ipsum dolor sit amet, consectetuer adipiscing elit.
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
