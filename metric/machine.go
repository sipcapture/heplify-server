//line machine.rl:1
package metric

import (
	"strings"
)

// ragel -G2 -Z machine.rl

//line machine.go:13
const metric_start int = 1
const metric_first_final int = 1
const metric_error int = 0

const metric_en_main int = 1

//line machine.rl:12

func extractXR(vq, data string) (s string) {
	if numPos := strings.Index(data, vq); numPos >= 0 {
		numPos += len(vq)
		data = data[numPos:]
	} else {
		return s
	}

	cs, p, pe, eof := 0, 0, len(data), len(data)
	mark := 0

//line machine.go:36
	{
		cs = metric_start
	}

//line machine.go:41
	{
		if p == pe {
			goto _test_eof
		}
		switch cs {
		case 1:
			goto st_case_1
		case 2:
			goto st_case_2
		case 3:
			goto st_case_3
		case 0:
			goto st_case_0
		}
		goto st_out
	st_case_1:
		switch data[p] {
		case 13:
			goto tr1
		case 32:
			goto tr1
		case 59:
			goto tr1
		}
		if 9 <= data[p] && data[p] <= 10 {
			goto tr1
		}
		goto tr0
	tr0:
//line machine.rl:26

		mark = p

		goto st2
	st2:
		if p++; p == pe {
			goto _test_eof2
		}
	st_case_2:
//line machine.go:81
		switch data[p] {
		case 13:
			goto tr3
		case 32:
			goto tr3
		case 59:
			goto tr3
		}
		if 9 <= data[p] && data[p] <= 10 {
			goto tr3
		}
		goto st2
	tr1:
//line machine.rl:26

		mark = p

//line machine.rl:30

		s = data[mark:p]

//line machine.rl:37
		return s
		goto st3
	tr3:
//line machine.rl:30

		s = data[mark:p]

//line machine.rl:37
		return s
		goto st3
	st3:
		if p++; p == pe {
			goto _test_eof3
		}
	st_case_3:
//line machine.go:119
		goto st0
	st_case_0:
	st0:
		cs = 0
		goto _out
	st_out:
	_test_eof2:
		cs = 2
		goto _test_eof
	_test_eof3:
		cs = 3
		goto _test_eof

	_test_eof:
		{
		}
		if p == eof {
			switch cs {
			case 2:
//line machine.rl:30

				s = data[mark:p]

			case 1:
//line machine.rl:26

				mark = p

//line machine.rl:30

				s = data[mark:p]

//line machine.go:146
			}
		}

	_out:
		{
		}
	}

//line machine.rl:41

	return s
}
