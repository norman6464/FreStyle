export { default as KnowledgeBaseRepository } from './api/knowledgeBaseRepository';
export {
  collectKbAncestorIds,
  replaceKbPageInTree,
  moveKbPageInTree,
  kbMoveActions,
  filterKbTree,
  collectKbBranchIds,
} from './lib/tree';
export type { KbDropTarget, KbMoveActions } from './lib/tree';
export type {
  KbWorkspace,
  KbSpace,
  KbPage,
  KbPageTreeNode,
  KbPageTree,
  KbPageDoc,
} from './model/types';
