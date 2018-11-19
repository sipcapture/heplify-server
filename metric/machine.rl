package metric

import (
  "strings"
)

// ragel -G2 -Z machine.rl

%%{
  machine metric;
  write data;
}%%

func extractXR(vq, data string) (s string) {
  if numPos := strings.Index(data, vq); numPos >= 0 {
      numPos += len(vq)
      data = data[numPos:]
  } else {
      return s
  }

  cs, p, pe, eof := 0, 0, len(data), len(data)
  mark := 0

  %%{
		action mark {
		  mark = p
		}

    action extract {
		  s = data[mark:p]
		}

    LWS = (' ' | ';' | '\n' | '\r'| '\t');
    
    num = any* >mark %extract;
    main := num :>LWS? @{ return s };

		write init;
		write exec;
	}%%

	return s
}