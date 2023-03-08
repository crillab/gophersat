package solver

const (
	nbMaxRecent      = 50 // How many recent LBD values we consider; "X" in papers about LBD.
	triggerRestartK  = 0.8
	nbMaxTrail       = 5_000 // How many elements in queueTrail we consider; "Y" in papers about LBD.
	postponeRestartT = 1.4
)

type queueData struct {
	totalNb   int     // Current total nb of values considered
	totalSum  int     // Sum of all values so far
	nbRecent  int     // NB of values used in the array
	ptr       int     // current index of oldest value in the array
	recentAvg float64 // Average value
}

// lbdStats is a structure dealing with recent LBD evolutions.
type lbdStats struct {
	lbdData      queueData
	trailData    queueData
	recentVals   [nbMaxRecent]int // Last LBD values
	recentTrails [nbMaxTrail]int  // Last trail lengths
}

// mustRestart is true iff recent LBDs are much smaller on average than average of all LBDs.
func (l *lbdStats) mustRestart() bool {
	if l.lbdData.nbRecent < nbMaxRecent {
		return false
	}
	return l.lbdData.recentAvg*triggerRestartK > float64(l.lbdData.totalSum)/float64(l.lbdData.totalNb)
}

// addConflict adds information about a conflict that just happened.
func (l *lbdStats) addConflict(trailSz int) {
	td := &l.trailData
	td.totalNb++
	td.totalSum += trailSz
	if td.nbRecent < nbMaxTrail {
		l.recentTrails[td.nbRecent] = trailSz
		old := float64(td.nbRecent)
		new := old + 1
		td.recentAvg = (td.recentAvg*old)/new + float64(trailSz)/new
		td.nbRecent++
	} else {
		old := l.recentTrails[td.ptr]
		l.recentTrails[td.ptr] = trailSz
		td.ptr++
		if td.ptr == nbMaxTrail {
			td.ptr = 0
		}
		td.recentAvg = td.recentAvg - float64(old)/nbMaxTrail + float64(trailSz)/nbMaxTrail
	}
	if td.nbRecent == nbMaxTrail && l.lbdData.nbRecent == nbMaxRecent && trailSz > int(postponeRestartT*td.recentAvg) {
		// Too many good assignments: postpone restart
		l.clear()
	}

}

// addLbd adds information about a recent learned clause's LBD.
// TODO: this is very close to addConflicts's code, this should probably be rewritten/merged.
func (l *lbdStats) addLbd(lbd int) {
	ld := &l.lbdData
	ld.totalNb++
	ld.totalSum += lbd
	if ld.nbRecent < nbMaxRecent {
		l.recentVals[ld.nbRecent] = lbd
		old := float64(ld.nbRecent)
		new := old + 1
		ld.recentAvg = (ld.recentAvg*old)/new + float64(lbd)/new
		ld.nbRecent++
	} else {
		old := l.recentVals[ld.ptr]
		l.recentVals[ld.ptr] = lbd
		ld.ptr++
		if ld.ptr == nbMaxRecent {
			ld.ptr = 0
		}
		ld.recentAvg = ld.recentAvg - float64(old)/nbMaxRecent + float64(lbd)/nbMaxRecent
	}
}

// clear clears last values. It should be called after a restart.
func (l *lbdStats) clear() {
	l.lbdData.ptr = 0
	l.lbdData.nbRecent = 0
	l.lbdData.recentAvg = 0.0
}
