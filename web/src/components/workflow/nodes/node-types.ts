import { TriggerNode } from "./trigger-node";
import { ActionNode } from "./action-node";
import { ConditionNode } from "./condition-node";

export const nodeTypes = {
  trigger: TriggerNode,
  action: ActionNode,
  condition: ConditionNode,
} as const;
