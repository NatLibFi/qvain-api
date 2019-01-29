// Package selfcheck runs some integrity checks on the configuration and operation of the application.
package selfcheck

// CheckWriter is an interface for different variations of output, such as console, html or json.
type CheckWriter interface {
	StartGroup(string)
	Tag(string)
	Msg(string)
	Good(string)
	Bad(string)
	Skipping()
	EndGroup()
}

// Run runs over the list of defined checks and outputs the results to the designated writer.
// It returns an int value indicating the number of errors encountered. This value can be passed directly to os.Exit() as exit code.
func Run(w CheckWriter) int {
	var nErrs int

	for _, grp := range checks {
		//fmt.Println("group:", grp.Name)
		w.StartGroup("")

		for _, chk := range grp.Checks {
			w.Tag(grp.Name)
			w.Msg(chk.Name)
			err := chk.Func()
			if err != nil {
				//fmt.Printf("[%02d] error: %s\n", i, err)
				nErrs++
				w.Bad(err.Error())
				w.Tag(grp.Name)
				w.Skipping()
				break
			} else {
				//fmt.Printf("[%02d] pass\n", i)
				w.Good("")
			}
		}
	}
	return nErrs
}
