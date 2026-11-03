import React from 'react';
import { Redirect } from '@docusaurus/router';

export default function DankCalendarTermsRedirect(): React.ReactNode {
  return <Redirect to="/dcal/terms" />;
}
