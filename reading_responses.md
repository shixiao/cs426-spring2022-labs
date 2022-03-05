# Reading Responses
"Writing is thinking. To write well is to think clearly." Writing responses is an opportunity to clarify and consolidate your thinking on the assigned reading.

## Logistics
* Materials (mostly papers) we'd like you to write reading responses for will be marked as "**PREREAD** for next class" in the [schedule](http://cs426.cloud/schedule.shtml) page.
* Write your reading responses in the google doc shared with you ("<NetID> CS426 Reading Responses"; check "Shared with me" on your google drive).
* Responses are due before class time for which the reading is assigned.

## Parameters
* Write **200-300 words** (ball park, not a hard limit) in reponse to the assigned reading. This should be _your_ critical review or commentary of the paper or the system.
* **Structure and content.**
  * Your response should not be a mere summary, though it is often helpful to start with a 1~2 sentence distillation of the main point or contribution of the paper / system.
  * Prioritize clarity over lavishness of prose, i.e., no need to wordsmith and focus on your arguments.
  * We are not looking for a particular right "answer" or "view" of the paper or system, but rather whether your stance is compelling and well supported.
* Some general prompts around which you can formulate your response, but you are not obligated to answer each of these prompts.
  * What are the authors' arguments / observations / insights / tradeoffs?
  * What worked well about the system? What would be the disadvantages?
  * Do you buy the authors' arguments? Why or why not?
  * Compare with one other system: consider systems that precede or come after the paper, what's similar? what's different? Why did the systems make different choices? To find related systems, it's often useful to look at the Related Work section of a paper or find on [Google Scholar](https://scholar.google.com/) later papers that cite the one at hand.
  * Some reading response assignments have paper-specific prompts or questions below that might also guide your reading.
* **Grading:** the goal of this assignment is to ensure you have done your reading and get a chance for independent thinking and writing. This will be closer to participation grade unless you submit something entirely incoherent or made-up.

## Specific prompts
These prompts are also questions for you to think about and not necessarily a checklist you must complete.

### Chubby
* What are the tradeoffs Chubby made in offering a coarse-grained locking API?
* Contrast DNS and Chubby's unexpected usage as a name server (Section 4.3).
* Which lessons and discussions surprised you? Why?

### Raft (Sections 6 and 7)
* Contrast the master failover mechanism in Chubby (Section 2.9) with the Raft leader election and cluster membership change protocol.

### Paxos
* Compare single-decree Paxos and Raft: in Raft, which part of the protocol acts as "learners"?
* How does one ensure the proposal number `n` is unique in Algorithm 12.1?

### Delos
* What are some of the ubiquitous examples of virtualization in other areas of computer science or related fields?
* What's the point of Delos if it still requires a MetaStore using Paxos or Raft? Why not directly use the Paxos- or Raft-based store?
* Are there scenarios where one would want to break out of the shared log API?

### Copysets
* What is the key observation that motivated Copysets?
* What is the main tradeoff that Copyset makes?
* Are Copyset choices coupled with / Do they conflict with failure domain awareness? Why or why not?

### TAO

### Weak consistency:
#### Dynamo (Sections 1, 2, 3.3, 4, 6-6.1, and 6.3)
* What is Dynamo's strategy for ensuring durability?
* Can you construct a scenario where the hinted handoff scheme results in an inconsistency?
* How would you evaluate and measure Dynamo's consistency guarantees?
* What properties should an application have to leverage Dynamo's read time conflict resolution?

#### Zanzibar (Section 2.2)
* Why wasn't Spanner (Google's linearizable store) used directly?
* How does the Zanzibar consistency protocol differ from performing direct reads and writes to Spanner? (Hint: what timestamp choice does Zanzibar have that Spanner has no knowledge of?)

#### FlightTracker (Sections 2 and 3)
* Compare and contrast FlightTracker Tickets and Zanzibar zookies.
* How would _you_ design Tickets and the Ticket store?

### ShardManager
