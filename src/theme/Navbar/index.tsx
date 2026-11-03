import React, { type ReactNode } from 'react';
import NavbarLayout from '@theme/Navbar/Layout';
import NavbarContent from './Content';

export default function NavbarWrapper(): ReactNode {
  return <NavbarLayout><NavbarContent /></NavbarLayout>;
}
