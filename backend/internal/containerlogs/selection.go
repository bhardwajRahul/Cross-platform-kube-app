package containerlogs

import "strings"

const (
	SelectedPodPrefix       = "pod:"
	SelectedInitPrefix      = "init:"
	SelectedContainerPrefix = "container:"
	SelectedDebugPrefix     = "debug:"
)

// ScopeSelection captures explicit pod/container selections from the object-panel logs UI.
// When present, these selections narrow the backend target set before per-scope/global caps.
type ScopeSelection struct {
	selectedPods       map[string]struct{}
	selectedContainers map[ContainerRef]struct{}
}

func ParseScopeSelection(values []string) ScopeSelection {
	selection := ScopeSelection{}
	for _, rawValue := range values {
		value := strings.TrimSpace(rawValue)
		selection.addValue(value)
	}
	return selection
}

func (s *ScopeSelection) addValue(value string) {
	if name, ok := selectedFilterName(value, SelectedPodPrefix); ok {
		s.addPod(name)
		return
	}
	if name, ok := selectedFilterName(value, SelectedInitPrefix); ok {
		s.addContainer(ContainerRef{Name: name, IsInit: true})
		return
	}
	if name, ok := selectedFilterName(value, SelectedDebugPrefix); ok {
		s.addContainer(ContainerRef{Name: name, IsEphemeral: true})
		return
	}
	if name, ok := selectedFilterName(value, SelectedContainerPrefix); ok {
		s.addContainer(ContainerRef{Name: name})
	}
}

func selectedFilterName(value, prefix string) (string, bool) {
	if !strings.HasPrefix(value, prefix) {
		return "", false
	}
	name := strings.TrimSpace(strings.TrimPrefix(value, prefix))
	return name, name != ""
}

func (s *ScopeSelection) addPod(name string) {
	if s.selectedPods == nil {
		s.selectedPods = make(map[string]struct{})
	}
	s.selectedPods[name] = struct{}{}
}

func (s *ScopeSelection) addContainer(container ContainerRef) {
	if s.selectedContainers == nil {
		s.selectedContainers = make(map[ContainerRef]struct{})
	}
	s.selectedContainers[container] = struct{}{}
}

func (s ScopeSelection) IsZero() bool {
	return len(s.selectedPods) == 0 && len(s.selectedContainers) == 0
}

func (s ScopeSelection) MatchPod(podName string) bool {
	if len(s.selectedPods) == 0 {
		return true
	}
	_, ok := s.selectedPods[podName]
	return ok
}

func (s ScopeSelection) MatchContainer(container ContainerRef) bool {
	if len(s.selectedContainers) == 0 {
		return true
	}
	_, ok := s.selectedContainers[container]
	return ok
}
