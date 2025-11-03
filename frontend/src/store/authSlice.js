import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  sub: null,
  name: null,
  email: null,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuthData(state, action) {
      const { sub, name, email } = action.payload;

      state.sub = sub;
      state.name = name;
      state.email = email;
    },

    clearAuthData(state) {
      state.sub = null;
      state.name = null;
      state.email = null;
    },
  },
});

export const { setAuthData, clearAuthData } = authSlice.actions;
export default authSlice.reducer;
