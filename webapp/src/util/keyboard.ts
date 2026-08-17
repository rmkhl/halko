import React from "react";

// Swallow the keys that activate a focused button, so a destructive control
// can only be triggered by a deliberate click or tap.
//
// A stray Enter is easy to land on a control that quietly holds focus, and
// Space doubles as the scroll key, so either can fire a button the user was
// not looking at. Attach this to the destructive button itself - never to a
// Dialog, which would swallow the keys for Cancel too, and never to Cancel or
// any other way out. Escape is left alone. The result is that an accidental
// keypress can only ever back out of a destructive action, never commit one.
//
// preventDefault on keydown is what stops Space: a native button fires its
// click on keyup as the default action of the keydown.
export const blockActivationKeys = (event: React.KeyboardEvent) => {
  if (event.key === "Enter" || event.key === " ") {
    event.preventDefault();
    event.stopPropagation();
  }
};
