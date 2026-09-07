export { default as KbRepository } from './api/kbRepository';
export { default as KbWorkspaceSwitcher } from './ui/KbWorkspaceSwitcher';
export type { KbWorkspaceSwitcherProps } from './ui/KbWorkspaceSwitcher';
export { useWorkspaceList } from './model/useWorkspaceList';
export { NOTE_NEW_PAGE_TITLE } from './config/constants';
export { subscribeKbTreeEvents, emitKbTreeEvent } from './model/kbTreeEvents';
export type { KbTreeEvent } from './model/kbTreeEvents';
export {
  collectKbAncestorIds,
  replaceKbPageInTree,
  moveKbPageInTree,
  kbMoveActions,
} from './lib/tree';
export type { KbDropTarget, KbMoveActions } from './lib/tree';
export { rememberVisitedPage, getLastVisitedPageId, forgetVisitedPageIfMatches } from './lib/lastVisitedPage';
export type {
  KbWorkspace,
  KbSpace,
  KbIcon,
  KbEditorRef,
  KbPage,
  KbPageTreeNode,
  KbPageTree,
  KbPageDoc,
  KbResolvedPage,
  KbResolvedCover,
  KbAncestorRef,
  KbGrantRole,
  KbPageGrant,
  KbGrantablePrincipal,
  KbCommentAuthorRef,
  KbComment,
  KbCommentThread,
  KbPageContentSaveResult,
  KbPageVersion,
  KbPageVersionDetail,
} from './model/types';
