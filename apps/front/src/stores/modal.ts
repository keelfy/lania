import { createStore } from 'zustand'
import { devtools } from 'zustand/middleware'

export enum ModalType {
  // default
  None = 'none',

  // orders
  ApproveOrder = 'approve-order',
  RejectOrder = 'reject-order',
  CreateOrder = 'create-order',

  // profile
  ProfileSearch = 'profile-search',
  ProfileSettings = 'profile-settings',

  // content notes
  DeleteContentAlert = 'delete-content-alert',
  ManualNoteCreation = 'manual-note-creation',
  GameNote = 'game-note',
  GameNoteEditor = 'game-note-editor',
  GameNoteCreator = 'game-note-creator',
  MovieNote = 'movie-note',
  MovieNoteEditor = 'movie-note-editor',
  MovieNoteCreator = 'movie-note-creator',

  // collections
  EditCollection = 'edit-collection',
  AddCollectionItem = 'add-collection-item',
  CreateCollection = 'create-collection',
  DeleteCollectionAlert = 'delete-collection-alert',

  // IGDB
  SelectContentItem = 'select-content-item',

  // show more profile
  ShowMoreProfile = 'show-more-profile',
}

export type ModalParamValue = string | string[] | undefined

type Action = {
  openModal: (
    modalName: ModalType,
    params?: Record<string, ModalParamValue>,
  ) => void
  setModalParams: (params: Record<string, ModalParamValue>) => void
  closeModal: () => void
}

type State = {
  currentModal: ModalType
  modalParams?: Record<string, ModalParamValue>
}

export type ModalStore = Action & State

const defaultInitialState: State = {
  currentModal: ModalType.None,
  modalParams: {},
}

const createModalStore = (initialState: State = defaultInitialState) => {
  return createStore<ModalStore>()(
    devtools((set) => ({
      ...initialState,
      openModal: (modalName, params) =>
        set(() => ({
          currentModal: modalName,
          modalParams: params ?? {},
        })),
      setModalParams: (params) =>
        set((state) => ({ ...state, modalParams: params })),
      closeModal: () => set(() => defaultInitialState),
    })),
  )
}

export default createModalStore
