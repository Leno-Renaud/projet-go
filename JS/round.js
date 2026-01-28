import { CARD_TYPES } from "./card.js";
import { ask } from "./prompt.js";

export default class Round {
  constructor(deck, players, logger) {
    this.deck = deck;
    this.players = players;
    this.logger = logger;
    this.finished = false;
    this.flip7Winner = null;
  }

  normalizeAction(raw) {
    const v = String(raw ?? "").trim().toLowerCase();
    if (["s", "stop", "rester", "stay"].includes(v)) return "s";
    if (["p", "piocher", "pioche", "draw", "d", "f", "flip"].includes(v)) return "p";
    return null;
  }

  async chooseTarget({ fromPlayer, effectLabel, allowSelf = true } = {}) {
    const eligible = allowSelf ? this.players : this.players.filter(p => p !== fromPlayer);

    // Si on ne peut pas choisir (ex: 1 seul joueur), fallback.
    if (eligible.length === 1) return eligible[0];

    console.log(`Choisir une cible pour ${effectLabel}:`);
    eligible.forEach((p, i) => {
      const flags = [
        p.active ? "actif" : "inactif",
        p.stayed ? "resté" : "",
        p.frozen ? "freeze" : ""
      ].filter(Boolean).join(", ");
      console.log(`  ${i + 1}) ${p.name}${flags ? ` (${flags})` : ""}`);
    });

    // eslint-disable-next-line no-constant-condition
    while (true) {
      const raw = await ask("Cible (numéro ou nom) : ");
      const byNum = Number(raw);
      if (Number.isInteger(byNum) && byNum >= 1 && byNum <= eligible.length) return eligible[byNum - 1];

      const byName = eligible.find(p => p.name.toLowerCase() === String(raw).toLowerCase());
      if (byName) return byName;

      console.log("Entrée invalide.");
    }
  }

  async play(roundNumber) {
    // reset état round
    this.roundOver = false;
    this.flip7Player = null;

    this.players.forEach(p => {
      p.resetRound();
      p.drawnCards = [];
    });

    this.logger.startRound(roundNumber, this.players);
    this.finished = false;
    this.flip7Winner = null;

    // distribution initiale
    for (const p of this.players) {
      await this.drawCard(p);
      if (this.roundOver) break;
    }

    while (!this.finished) {
      for (const p of this.players.filter(x => x.active && !x.stayed)) {
        if (this.roundOver) break;

        console.log(`\n${p.name} a pioché:`, p.drawnCards.map(c => this.formatCard(c)).join(", "));

        let choice = null;
        while (!choice) {
          const choiceRaw = await ask(`${p.name} → (p)iocher ou (s)rester ? `);
          choice = this.normalizeAction(choiceRaw);
          if (!choice) console.log("Entrée invalide (utiliser p/s)." );
        }

        if (choice === "s") {
          p.stayed = true;
          this.logger.log({ type: "stay", player: p.name });
          continue;
        }

        await this.drawCard(p);
        if (this.finished) break;
      }

      if (this.finished) break;

      const activeLeft = this.players.some(p => p.active && !p.stayed);
      if (!activeLeft) break;
    }

    if (this.flip7Player) {
      console.log(`${this.flip7Player.name} a fait FLIP 7 !`);
      this.logger.log({ type: "flip7", player: this.flip7Player.name });
    }

    // scoring UNE SEULE FOIS, fin de round
    this.scoreRound();

    this.logger.endRound(this.players);
  }

  formatCard(card) {
    if (!card) return "Aucune carte";
    return card.value !== undefined ? `${card.type}(${card.value})` : `${card.type}`;
  }

  scoreRound() {
    for (const p of this.players) {
      const score = p.scoreRound();
      p.totalScore += score;

      console.log(`${p.name} gagne ${score} (total ${p.totalScore})`);
      this.logger.log({ type: "roundScore", player: p.name, score, totalScore: p.totalScore });
    }
  }

  async drawCard(player) {
    if (this.roundOver) return;

    const card = this.deck.draw();
    this.logger.log({ type: "draw", player: player.name, card });

    if (!card) return;

    // Stocke la carte dans l'historique du joueur (cartes "gardées" pour l'affichage)
    if (!player.drawnCards) player.drawnCards = [];
    player.drawnCards.push(card);

    console.log(`${player.name} pioche: ${this.formatCard(card)}`);

    switch (card.type) {
      case CARD_TYPES.NUMBER: {
        if (player.hasDuplicate(card.value)) {
          if (player.secondChance) {
            player.secondChance = false;
            // On défausse le doublon + on consomme la carte SECOND_CHANCE (retirée de l'affichage)
            const last = player.drawnCards[player.drawnCards.length - 1];
            if (last && last.type === CARD_TYPES.NUMBER && last.value === card.value) player.drawnCards.pop();
            const idx = [...player.drawnCards].reverse().findIndex(c => c.type === CARD_TYPES.SECOND_CHANCE);
            if (idx !== -1) {
              const realIndex = player.drawnCards.length - 1 - idx;
              player.drawnCards.splice(realIndex, 1);
            }

            this.logger.log({ type: "second_chance_used", player: player.name, duplicate: card.value });
            console.log(`${player.name}: second chance utilisée (doublon ${card.value} défaussé).`);
            break;
          }
          console.log("Doublon → éliminé !");
          player.active = false;
        } else {
          player.addNumber(card.value);

          if (player.numbers.length === 7) {
            player.stayed = true; // encaissement automatique
            this.finished = true;
            this.flip7Winner = player;
            this.logger.log({ type: "flip7", player: player.name });
            console.log(`${player.name} a fait FLIP 7 !`);
          }
        }
        break;
      }

      case CARD_TYPES.FREEZE: {
        const activePlayers = this.players.filter(p => p.active && p !== player);
        
        if (activePlayers.length === 0) {
          console.log("Aucun autre joueur actif !");
          break;
        }

        console.log("\nJoueurs actifs disponibles :");
        activePlayers.forEach((p, index) => {
          console.log(`${index + 1}. ${p.name}`);
        });

        let selectedIndex = -1;
        while (selectedIndex < 0 || selectedIndex >= activePlayers.length) {
          const input = await ask(`${player.name}, choisis un joueur (1-${activePlayers.length}) : `);
          selectedIndex = parseInt(input) - 1;
        }

        const frozenPlayer = activePlayers[selectedIndex];
        frozenPlayer.stayed = true;
        console.log(`${frozenPlayer.name} finit son tour et marque ses points !`);
        this.logger.log({ type: "freeze", player: player.name, targetPlayer: frozenPlayer.name });
        break;
      }

      case CARD_TYPES.FLIP_THREE:
        for (let i = 0; i < 3; i++) {
          if (this.roundOver) break;
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