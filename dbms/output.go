package dbms

import (
	"database/sql"
	go_ora "github.com/sijms/go-ora/v2"
	"io"
)

type DBOutput struct {
	bufferSize int
	conn       *sql.DB
}

func NewOutput(conn *sql.DB, bufferSize int) (*DBOutput, error) {
	output := &DBOutput{
		bufferSize: bufferSize,
		conn:       conn,
	}
	sqlText := `begin dbms_output.enable(:1); end;`
	if output.bufferSize > 0x7FFF {
		output.bufferSize = 0x7FFF
	}
	if output.bufferSize < 2000 {
		output.bufferSize = 2000
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
	_, err := db_out.conn.Exec(sqlText, 0x7FFF, go_ora.Out{Dest: &state},
		go_ora.Out{Dest: &output, Size: db_out.bufferSize})
	return output, err
}

func (output *DBOutput) Close() error {
	_, err := output.conn.Exec(`begin dbms_output.disable; end;`)
	return err
}
