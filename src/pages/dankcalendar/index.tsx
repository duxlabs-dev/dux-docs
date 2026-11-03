import React from 'react';
import { Redirect } from '@docusaurus/router';

export default function DankCalendarRedirect(): React.ReactNode {
  return <Redirect to="/dcal" />;
}
