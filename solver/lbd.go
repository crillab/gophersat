package solver

const (
	nbMaxRecent     = 50 // How many recent LBD values we consider
	triggerRestartK = 0.8
)

// lbdStats is a structure dealing with recent LBD evolutions.
type lbdStats struct {
	totalNb    int              // Total number of values considered
	totalSum   int              // Sum of all LBD so far
	nbRecent   int              // Nb of values useful in recentVals
	recentVals [nbMaxRecent]int // Last LBD values
	ptr        int              // Current index of oldest value in recentVals
	recentAvg  float64          // Average LBD for recentVals
}

// mustRestart is true iff recent LBDs are much smaller on average than average of all LBDs.
func (l *lbdStats) mustRestart() bool {
	if l.nbRecent < nbMaxRecent {
		return false
	}
	return l.recentAvg*triggerRestartK > float64(l.totalSum)/float64(l.totalNb)
}

// add adds information about a recent learned clause's LBD.
func (l *lbdStats) add(lbd int) {
	l.totalNb++
	l.totalSum += lbd
	if l.nbRecent < nbMaxRecent {
		l.recentVals[l.nbRecent] = lbd
		oldNbRecent := float64(l.nbRecent)
		newNbRecent := float64(l.nbRecent + 1)
		l.recentAvg = (l.recentAvg*oldNbRecent)/newNbRecent + float64(lbd)/newNbRecent
		l.nbRecent++
	} else {
		oldVal := l.recentVals[l.ptr]
		l.recentVals[l.ptr] = lbd
		l.ptr++
		if l.ptr == nbMaxRecent {
			l.ptr = 0
		}
		l.recentAvg = l.recentAvg - float64(oldVal)/nbMaxRecent + float64(lbd)/nbMaxRecent
	}
}

// clear clears last values. It should be called after a restart.
func (l *lbdStats) clear() {
	l.ptr = 0
	l.nbRecent = 0
	l.recentAvg = 0.0
}
