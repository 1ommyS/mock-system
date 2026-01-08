import type { Meta, StoryObj } from "@storybook/react";
import { Input } from "./input";

const meta: Meta<typeof Input> = {
  title: "UI/Input",
  component: Input,
  args: {
    placeholder: "Введите email",
  },
};

export default meta;

type Story = StoryObj<typeof Input>;

export const Text: Story = {};

export const Password: Story = {
  args: {
    type: "password",
    placeholder: "Введите пароль",
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
    value: "Недоступно",
  },
};
