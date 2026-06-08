package picker

import (
	"bufio"
	"fmt"
	"os"

	"github.com/wstevenson1/favs/internal/store"
)

func Pick(commands []store.Command) (string, error) {
	if len(commands) == 0 {
		fmt.Fprintln(os.Stderr, "no saved commands. use 'favs add' to get started")
		return "", nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("could not open terminal: %w", err)
	}
	defer tty.Close()

	fmt.Fprint(tty, FormatList(commands))

	scanner := bufio.NewScanner(tty)
	for attempts := 0; attempts < 2; attempts++ {
		fmt.Fprintf(tty, "\nSelect (1-%d, q to quit): ", len(commands))
		if !scanner.Scan() {
			return "", nil
		}
		n, err := ParseSelection(scanner.Text(), len(commands))
		if err == ErrQuit {
			return "", nil
		}
		if err == ErrInvalid {
			if attempts == 0 {
				fmt.Fprintf(tty, "Invalid selection. Try again.\n")
				continue
			}
			return "", nil
		}
		return commands[n-1].Command, nil
	}
	return "", nil
}
