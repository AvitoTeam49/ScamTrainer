package scenario

import (
	"errors"
	"fmt"
)

func Validate(scenario *Scenario) error {

	if scenario == nil {
		return errors.New("scenario is nil")
	}

	if err := validateFields(scenario); err != nil {
		return err
	}

	for nodeID, node := range scenario.Nodes {
		if err := validateNode(scenario, nodeID, node); err != nil {
			return err
		}
	}

	if err := validateDAG(scenario); err != nil {
		return fmt.Errorf("validate graph: %w", err)
	}

	return nil
}

func validateFields(scenario *Scenario) error {

	if scenario.ID == "" {
		return errors.New("scenario id is empty")
	}

	if scenario.Title == "" {
		return errors.New("scenario title is empty")
	}

	if scenario.Role != RoleBuyer && scenario.Role != RoleSeller {
		return fmt.Errorf("unsupported role %q", scenario.Role)
	}

	if len(scenario.Nodes) == 0 {
		return errors.New("scenario contains no nodes")
	}

	if _, exists := scenario.Nodes[scenario.StartNodeID]; !exists {
		return fmt.Errorf(
			"start node %q does not exist",
			scenario.StartNodeID,
		)
	}
	return nil
}

func validateNode(scenario *Scenario, nodeID string, node *Node) error {
	if node == nil {
		return fmt.Errorf("node %q is nil", nodeID)
	}

	switch node.Type {
	case
		NodeTypeDecision, NodeTypeEnding:
	default:
		return fmt.Errorf(
			"node %q has unsupported type %q",
			nodeID,
			node.Type,
		)
	}

	if node.Type == NodeTypeEnding {
		if len(node.Transitions) != 0 {
			return fmt.Errorf(
				"ending node %q must not contain transitions",
				nodeID,
			)
		}

		return nil
	}

	if len(node.Transitions) == 0 {
		return fmt.Errorf(
			"non-ending node %q must contain transitions",
			nodeID,
		)
	}

	transitionIDs := make(map[string]struct{})

	for _, transition := range node.Transitions {
		if transition.ID == "" {
			return fmt.Errorf(
				"node %q contains transition without id",
				nodeID,
			)
		}

		if _, exists := transitionIDs[transition.ID]; exists {
			return fmt.Errorf(
				"node %q contains duplicate transition %q",
				nodeID,
				transition.ID,
			)
		}

		transitionIDs[transition.ID] = struct{}{}

		if transition.ToNodeID == "" {
			return fmt.Errorf(
				"transition %q from node %q has empty target",
				transition.ID,
				nodeID,
			)
		}

		if _, exists := scenario.Nodes[transition.ToNodeID]; !exists {
			return fmt.Errorf(
				"transition %q from node %q references missing node %q",
				transition.ID,
				nodeID,
				transition.ToNodeID,
			)
		}
	}

	return nil
}

type visitState uint8

const (
	notVisited visitState = iota
	visiting
	visited
)

func validateDAG(scenario *Scenario) error {
	states := make(
		map[string]visitState,
		len(scenario.Nodes),
	)

	var visit func(nodeID string) error

	visit = func(nodeID string) error {
		switch states[nodeID] {
		case visiting:
			return fmt.Errorf(
				"cycle detected at node %q",
				nodeID,
			)

		case visited:
			return nil
		}

		states[nodeID] = visiting

		node := scenario.Nodes[nodeID]

		for _, transition := range node.Transitions {
			if err := visit(transition.ToNodeID); err != nil {
				return fmt.Errorf(
					"transition %q from node %q to node %q: %w",
					transition.ID,
					nodeID,
					transition.ToNodeID,
					err,
				)
			}
		}

		states[nodeID] = visited

		return nil
	}

	if err := visit(scenario.StartNodeID); err != nil {
		return err
	}

	for nodeID := range scenario.Nodes {
		if states[nodeID] == notVisited {
			return fmt.Errorf(
				"node %q is unreachable from start node %q",
				nodeID,
				scenario.StartNodeID,
			)
		}
	}
	return nil
}
