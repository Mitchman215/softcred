# US Bank Triple Cash $100 Software Credit — Research Notes

Last updated: 2026-04-21.

## Sources

- [US Bank — Triple Cash card page (footnote 3 has official credit terms)](https://www.usbank.com/business-banking/business-credit-cards/business-triple-cash-back-credit-card.html)
- [Doctor of Credit — What Works & Profitable Uses (article + 480+ comments)](https://www.doctorofcredit.com/u-s-bank-triple-cash-100-software-credit-what-works-profitable-uses/)
- [Doctor of Credit — Card overview](https://www.doctorofcredit.com/u-s-bank-triple-cash-rewards-business-card-500-signup-bonus-100-software-credit-3-categories-no-annual-fee/)
- [FlyerTalk — Auto-billing software transactions thread](https://www.flyertalk.com/forum/credit-card-programs/2083345-us-bank-triple-cash-rewards-bonus-auto-billing-software-transactions.html)
- [myFICO — $100 statement credit discussion](https://ficoforums.myfico.com/t5/Business-Credit/100-statement-credit-on-US-bank-Business-card/td-p/6706440)
- [US Bank — Program rules](https://rewards.usbank.com/benefits/card/BizTripleCashVBenefits/program-rules)

## How the Credit Works

- Make a purchase coded MCC 5734 ("Computer Software Stores") every month for **11 consecutive months**
- US Bank automatically applies a **$100 statement credit** within 1-2 billing cycles after the 11th month
- The credit is a single lump sum, not per-transaction reimbursement
- Resets on cardmember anniversary year (not calendar year)
- Amount of each purchase doesn't matter — $1.50/mo works the same as $50/mo
- Must be monthly recurring — annual subscriptions do **not** qualify
- "Account must be in good standing (open and able to use) to receive the credit"
- The credit reduces account balance but does **not** count as a payment

## Eligibility Email

After your first qualifying MCC 5734 purchase, US Bank may send an email:

> Subject: "You bought the software. Remember to earn $100 from it!"

Not everyone receives this email. Per multiple DPs, not receiving it doesn't mean anything — the credit still tracks and posts.

## Progress Tracker

- Available in the **Android** US Bank app
- **Not** available in the iOS app as of early 2026
- Not available on the web portal

## Checking MCC Codes

- Download transactions as CSV from US Bank online banking
- MCC is in the Memo column, second semicolon-delimited field (zero-padded 5 digits, e.g. `05734`)
- Example memo: `24011346103100126810835; 05734; ; ; ;`

## Timing Data Points

| User | Service | 11th Month | Credit Posted | Days |
|------|---------|-----------|--------------|------|
| OncRN | 1Password | 2026-01-07 | 2026-02-24 | 48 |
| Melinda | seats.aero | ~2026-02 | 2026-03-11 | ~30 |
| Churnlark | Darkhorseodds | ~2025-07 | 2025-08-18 | ~30 |
| Gene | DealCheck | ~2025-11 | 2025-12-10 | ~30 |
| the other Justin | unknown | ~2025-08 | 2025-09-10 | ~30 |
| Melon | Adobe | 11 months | "very quickly" | — |

## Confirmed Working Services (MCC 5734)

| Service | Monthly Cost | Notes |
|---------|-------------|-------|
| SimpleFIN Bridge | $1.50 | Cheapest known option. No FTF. |
| 1Password | $3.99 → $4.99 (Mar 2026) | FTF charged (Canada-based) but USB may reimburse on request |
| Bitwarden (Business) | $5.00 | Monthly only on business plan. **Stopped coding correctly at some point per some DPs** |
| MaxMyPoint | $3.99 | |
| seats.aero | $9.99 | |
| Cloudflare | varies | |
| Fastmail | varies | |
| ChatGPT/OpenAI | $5+ (API) | Manual API credits work; need monthly charge |
| Claude AI (Anthropic) | $20.00 | |
| Proton Mail | varies | Mixed reports — verify MCC |
| NextDNS | $1.99 | |
| Mullvad VPN | ~$5.00 | **Caution**: one user reported 13 qualifying payments with no credit — USB investigating |
| DealCheck | varies | |
| Darkhorseodds | varies | |
| Norton AV (monthly) | varies | |
| PointsPath | varies | |
| Kagi Search | varies | |
| Midjourney | varies | |
| FreshBooks | varies | Official USB example |
| QuickBooks | varies | Official USB example |
| Ninite | varies | |
| Inoreader | varies | |
| CubeBackup | varies | |
| Brave Search Premium | varies | |

## Confirmed NOT Working

| Service | MCC | Notes |
|---------|-----|-------|
| GitHub | 07372 | Computer Programming, Data Processing |
| Google One/Workspace/Domains | various | |
| Microsoft 365 | various | |
| Backblaze | varies | |
| Canva | varies | |
| Discord Nitro | varies | |
| Dropbox | varies | |
| Apple iCloud+ | 05818 | |
| Todoist | varies | |
| Namecheap | varies | |
| Simplelogin | varies | |

## MCC Drift Warnings

Services that have changed MCC codes unexpectedly:

### Adobe Creative Cloud (HIGH RISK)
- Previously coded consistently as 5734
- Starting ~April 2025, began coding as 5817, 4816, 5818 randomly
- One user had 3 years of successful credits, then Adobe changed
- Multiple independent reports confirm this — **do not rely on Adobe**

### CrashPlan
- Changed merchant processor → now codes as 5817 (Digital Goods)

### Bitwarden
- Some reports of it stopping to code as 5734
- "Wasted 10 months and $50 on it" — one user

## Known Issues & Edge Cases

### Missed month resets the counter
Even one month where the charge codes to a different MCC resets the 11-month counter to zero. This is the core risk.

### Multiple cards
You can have multiple Triple Cash cards and earn $100 on each independently. Multiple DPs confirm this strategy works.

### Multiple subscriptions on one card
Only one $100 credit per 12-month period per card, regardless of how many qualifying subscriptions you have.

### Foreign Transaction Fees
- Some services (1Password, BunnyCDN) trigger 3% FTF even when billed in USD, because the merchant is based outside the US
- Some users report USB reimbursing the FTF on request

### Small Balance Write-Off Trick
Pay card down to $0.99 before statement closing date → USB writes off the balance → $0 statement. Unrelated to software credit but commonly discussed alongside it.

## What the $100 Credit Looks Like on Statement

**Unknown.** No one in any forum has documented the exact transaction description text as it appears in CSV/statement data. We know:
- It's a CREDIT transaction (not a payment)
- Amount is exactly $100.00
- Appears 1-2 billing cycles after 11th qualifying month
- Does not have "PAYMENT THANK YOU" or "INTERNET PAYMENT" in the description

This will be identified once the first credit posts on one of our monitored cards.

## Official Terms

### From USB card page ([footnote 3](https://www.usbank.com/business-banking/business-credit-cards/business-triple-cash-back-credit-card.html))

> An automatic statement credit of $100 per 12-month period will be applied to your account within 1-2 statement billing cycles following 11 consecutive months of eligible software service purchases made directly with a software service provider. Eligible software service providers are identified by their Merchant Category Code (MCC) and purchases made at discount/retail stores or online retailers may not qualify. We reserve the right to adjust or reverse any portion or all of any software services credit for unauthorized purchases or transaction credits.

### From USB eligibility email

> An automatic statement credit of $100 per 12‑month period will be applied to your Account within two (2) statement billing cycles following 11 consecutive months of eligible software service purchases made directly with a software service provider. Eligible software service providers are identified by their Merchant Category Code (MCC) and purchases made at discount/retail stores or online retailers may not qualify. We do not determine the category codes that merchants choose and reserve the right to determine which Purchases qualify. We reserve the right to adjust or reverse any portion or all of any software services credit for unauthorized purchases or transaction credits. Account must be in good standing (open and able to use) to receive the credit.

Note: the email version includes two additional clauses not on the card page: "We do not determine the category codes that merchants choose" and "Account must be in good standing (open and able to use) to receive the credit."
