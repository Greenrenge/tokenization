# all of this can be the a pipeline in the transporter transfer mongo --> listening es the use terms aggs (but it makes more storage used) 

# a go routine can handle all of processing 
# receive chunk of works [1000,1000,1000] from message queue ( batch no 1 of 300 of key xxxx)

## how it knows when "its ending"
## keep chunk key to mongo and uses its atomic operation to decide that kinda job is done or not, { _id:"",total,batches : { 1 : ts , 2 :ts }} ,read then writes, if fail, them try until it is full of batch.  the last full filled batch no will responsible to trigger other worker pool routines to make summary and make a result of 200 words and post its result to alert-graph api ??

1) char filter remove @ # <> emoji ๆ ! -->''

เราคิดว่าประเทศไทยเนี่ย ดีที่สุดเลยนะ 5555 ไปๆๆๆๆ ไปเที่ยวไทยกัน #thai 555 

2.1 phrase splitter (\n \s+ )

2.2) phrase selection (thai) desire sentence set ( all, expression set,noun by dict ) == ahocorasik  , link rejection , reject alone stopword 
ดีที่สุด ( ดี แย่ น่ารัก ห่วย กาก other words that express feeling of people)

3) tokenization (by space or by lib (tha jpn) )
เรา คิด ว่า ประเทศ ไทย เนี่ย  ดี ที่สุด เลย นะ 5555 ไป ไปเที่ยวไทยกัน

4) stemming 
presentation --> present
๒-->2
ร้าก-->รัก
ฉัน-->ชั้น

5) token filter remove repeating pattern
ฮือ (ออออ) (5555)

6) detokenization nGram
เราคิด เราคิดว่า


desire grams set ( all, expression set,  ) == ahocorasik