package dbms

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	go_ora "github.com/sijms/go-ora/v2"
)

type DBOutput struct {
	bufferSize int
	conn       *sql.DB
}

const (
	MaxBufferSize = 0x7FFF
	MinBufferSize = 2000
	KeyInContext  = "GO-ORA.DBMS_OUTPUT"
)

// enable oracle output for current session
// param:
//
//	ctx: context of goroutine used in large apps
//	     for main: context.Background()
//	     for rest apis:
//	       http.Request.Context()
//	       gin.Context
//	       fiber.Ctx.Context()
//	       ...
func EnableOutput(ctx context.Context, conn *sql.DB) error {
	out, err := NewOutput(conn, MaxBufferSize)
	if err != nil {
		return err
	}
	context.WithValue(ctx, KeyInContext, out)
	return nil
}

// disable oracle output for current session
func DisableOutput(ctx context.Context) error {
	out := ctx.Value(KeyInContext)
	if out == nil {
		return fmt.Errorf("invalid context")
	}
	err := out.(*DBOutput).Close()
	if err != nil {
		return err
	}
	return nil
}

// get oracle output for current session
func GetOutput(ctx context.Context) (string, error) {
	out := ctx.Value(KeyInContext)
	if out == nil {
		return "", fmt.Errorf("invalid context")
	}
	output, err := out.(*DBOutput).GetOutput()
	if err != nil {
		return "", err
	}
	return output, nil
}

// print oracle output into StringWriter for current session
func PrintOutput(ctx context.Context, w io.StringWriter) error {
	output, err := GetOutput(ctx)
	if err != nil {
		return err
	}
	_, err = w.WriteString(output)
	return err
}

func NewOutput(conn *sql.DB, bufferSize int) (*DBOutput, error) {
	output := &DBOutput{
		bufferSize: bufferSize,
		conn:       conn,
	}
	sqlText := `begin dbms_output.enable(:1); end;`
	if output.bufferSize > MaxBufferSize {
		output.bufferSize = MaxBufferSize
	}
	if output.bufferSize < MinBufferSize {
		output.bufferSize = MinBufferSize
	}
	_, err := output.conn.Exec(sqlText, bufferSize)
	return output, err
}

func (db_out *DBOutput) Print(w io.StringWriter) error {
	line, err := db_out.GetOutput()
	if err != nil {
		return err
	}
	_, err = w.WriteString(line)
	return err
}
func (db_out *DBOutput) GetOutput() (string, error) {
	sqlText := `declare 
	l_line varchar2(255); 
	l_done number; 
	l_buffer long; 
begin 
 loop 
 exit when length(l_buffer)+255 > :maxbytes OR l_done = 1; 
 dbms_output.get_line( l_line, l_done ); 
 if length(l_line) > 0 then
 	l_buffer := l_buffer || l_line || chr(10); 
 end if;
 end loop; 
 :done := l_done; 
 :buffer := l_buffer; 
end;`
	var (
		state  int
		output string
	)
	_, err := db_out.conn.Exec(sqlText, MaxBufferSize, go_ora.Out{Dest: &state},
		go_ora.Out{Dest: &output, Size: db_out.bufferSize})
	return output, err
}

func (output *DBOutput) Close() error {
	_, err := output.conn.Exec(`begin dbms_output.disable; end;`)
	return err
}
