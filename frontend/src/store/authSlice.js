import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  sub: localStorage.getItem('sub') || null,
  name: localStorage.getItem('name') || null,
  email: localStorage.getItem('email') || null,
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

      localStorage.setItem('sub', sub);
      localStorage.setItem('name', name);
      localStorage.setItem('email', email);
    },

    clearAuthData(state) {
      state.sub = null;
      state.name = null;
      state.email = null;

      localStorage.removeItem('sub');
      localStorage.removeItem('name');
      localStorage.removeItem('email');
    },
  },
});

export const { setAuthData, clearAuthData } = authSlice.actions;
export default authSlice.reducer;
