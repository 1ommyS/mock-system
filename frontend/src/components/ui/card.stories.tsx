import type { Meta, StoryObj } from "@storybook/react";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "./card";
import { Button } from "./button";

const meta: Meta<typeof Card> = {
  title: "UI/Card",
  component: Card,
};

export default meta;

type Story = StoryObj<typeof Card>;

export const Default: Story = {
  render: () => (
    <Card className="max-w-md">
      <CardHeader>
        <CardTitle>Liquid Glass Card</CardTitle>
        <CardDescription>
          Компонент карточки с мягким стеклянным эффектом.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-[var(--text-muted)]">
          Используйте для форм, уведомлений и статуса.
        </p>
      </CardContent>
      <CardFooter>
        <Button>Продолжить</Button>
      </CardFooter>
    </Card>
  ),
};
