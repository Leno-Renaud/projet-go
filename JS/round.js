import { CARD_TYPES } from "./card.js";
import { ask } from "./prompt.js";

export default class Round {
  constructor(deck, players, logger) {
    this.deck = deck;
    this.players = players;
    this.logger = logger;
  }

  async play(roundNumber) {
    this.players.forEach(p => p.resetRound());
    this.logger.startRound(roundNumber, this.players);

    // distribution initiale
    for (const p of this.players) {
      await this.drawCard(p);
    }

    let finished = false;

    while (!finished) {
      for (const p of this.players.filter(x => x.active && !x.stayed)) {
        const choice = await ask(
          `\n${p.name} → (p)iocher ou (s)rester ? `
        );

        if (choice === "s") {
          p.stayed = true;
          this.logger.log({ type: "stay", player: p.name });
          continue;
        }

        await this.drawCard(p);

        if (p.numbers.length === 7) {
          console.log(`${p.name} a fait FLIP 7 !`);
          finished = true;
          break;
        }
      }

      const activeLeft = this.players.some(p => p.active && !p.stayed);
      if (!activeLeft) break;
    }

    // scoring
    for (const p of this.players) {
      const score = p.scoreRound();
      p.totalScore += score;
      console.log(`${p.name} gagne ${score} (total ${p.totalScore})`);
    }

    this.logger.endRound(this.players);
  }

  async drawCard(player) {
    const card = this.deck.draw();
    this.logger.log({ type: "draw", player: player.name, card });

    if (!card) return;

    switch (card.type) {
      case CARD_TYPES.NUMBER:
        if (player.hasDuplicate(card.value)) {
          if (player.secondChance) {
            player.secondChance = false;
            console.log("Second chance utilisée !");
            break;
          }
          console.log("Doublon → éliminé !");
          player.active = false;
        } else {
          player.addNumber(card.value);
        }
        break;

      case CARD_TYPES.FREEZE:
        player.frozen = true;
        player.active = false;
        break;

      case CARD_TYPES.FLIP_THREE:
        for (let i = 0; i < 3; i++) {
          await this.drawCard(player);
        }
        break;

      case CARD_TYPES.SECOND_CHANCE:
        player.secondChance = true;
        break;

      case CARD_TYPES.BONUS:
        player.bonuses.push(card.value);
        break;

      case CARD_TYPES.MULTIPLIER:
        player.multiplier = true;
        break;
    }
  }
}
